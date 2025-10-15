-- name: GetActiveUserServiceRequests :many
SELECT
    sr.*,
    requester.full_name AS requester_name,
    provider.full_name  AS provider_name,
    l.title
FROM
    service_request sr
JOIN "user" requester ON sr.requester_id = requester.id
JOIN "user" provider  ON sr.provider_id = provider.id
JOIN service_listing l on sr.listing_id  = l.id
WHERE
    (sr.provider_id = $1 OR sr.requester_id = $1)
    AND sr.activity = 'active';




-- name: GetRequestByID :one
SELECT
  sr.id AS sr_id,
  sr.listing_id AS sr_listing_id,
  sr.status_detail AS sr_status_detail,
  sr.activity AS sr_activity,
  sr.created_at AS sr_created_at,
  sr.updated_at AS sr_updated_at,
  sr.token_reward AS sr_token_reward,

  sl.id AS sl_id,
  sl.title AS sl_title,
  sl.description AS sl_description,
  sl.posted_by AS sl_posted_by,
  sl.posted_at AS sl_posted_at,
  sl.category AS sl_category,

  ru.id AS requester_id,
  ru.full_name AS requester_full_name,
  ru.joined_at AS requester_joined_at,

  pu.id AS provider_id,
  pu.full_name AS provider_full_name,
  pu.joined_at AS provider_joined_at,
  CASE 
    WHEN rr.status = 'ongoing' THEN TRUE
    ELSE FALSE
  END AS ticket_open,
  COALESCE(sc.requester_completed,false),
  COALESCE(sc.provider_completed,false),
  COALESCE(
    json_agg(
      json_build_object(
        'event_id', e.id,
        'event_time', e.created_at,
        'event_description', e.description,
        'event_owner', n.action_user_id
      )
      ORDER BY e.created_at
    ) FILTER (WHERE e.id IS NOT NULL),
    '[]'::json
  )::json AS events
FROM service_request sr
JOIN service_listing sl ON sr.listing_id = sl.id
JOIN "user" ru ON sr.requester_id = ru.id
JOIN "user" pu ON sr.provider_id = pu.id
LEFT JOIN service_request_completion sc ON sr.id = sc.request_id
LEFT JOIN "event" e ON e.target_id = sr.id
LEFT JOIN notification n
ON n.event_id = e.id
LEFT JOIN request_report rr
ON rr.request_id = sr.id
WHERE sr.id = $1
GROUP BY
  sr.id, sl.id, ru.id, pu.id, sc.requester_completed, sc.provider_completed,rr.id;

-- name: InsertPendingServiceRequest :one
INSERT INTO service_request (listing_id,requester_id,provider_id,status_detail,activity,created_at,updated_at,token_reward)
SELECT
    $1,
    $2,
    sl.posted_by,
    'pending', 'active', NOW(),NOW(),sl.token_reward
FROM service_listing sl
WHERE sl.id = $1 AND sl.posted_by != $2
RETURNING id;


-- name: UpdateServiceRequest :one
UPDATE service_request
SET status_detail = $1, activity = $2, updated_at = NOW()
WHERE id = $3
RETURNING id;

-- name: InsertServiceRequestCompletion :exec
INSERT INTO service_request_completion (request_id,requester_completed,provider_completed,is_active)
VALUES ($1,false,false,true);

-- name: GetServiceRequestCompletion :one
SELECT * FROM service_request_completion
WHERE request_id = $1;


-- name: UpdateServiceRequestCompletion :exec
UPDATE service_request_completion
SET requester_completed = $1, provider_completed = $2, is_active = $3
WHERE request_id = $4;

-- name: GetAllUserRequests :many
SELECT
    sr.*,
    requester.full_name AS requester_name,
    provider.full_name  AS provider_name,
    l.title
FROM
    service_request sr
JOIN "user" requester ON sr.requester_id = requester.id
JOIN "user" provider  ON sr.provider_id = provider.id
JOIN service_listing l on sr.listing_id  = l.id
WHERE
    (sr.provider_id = sqlc.arg(user_id) OR sr.requester_id = sqlc.arg(user_id));




-- name: InsertRequestReport :one
INSERT INTO request_report (reporter_id, request_id, ticket_id, created_at,updated_at, "status")
VALUES ($1, $2, '', NOW(),NOW(), 'ongoing')
RETURNING id, created_at;

-- name: UpdateRequestReportWithTicketID :one
UPDATE request_report
SET ticket_id = $1
WHERE id = $2
RETURNING ticket_id;


-- name: GetRequestReport :one
SELECT * FROM request_report
WHERE request_id = $1 AND reporter_id = $2;


-- name: UpdateExpiredRequest :many
UPDATE service_request AS sr
SET status_detail = 'expired', activity = 'inactive', updated_at = NOW()
FROM service_listing AS sl
JOIN "user" AS ru ON ru.id = sr.requester_id
WHERE sl.id = sr.listing_id AND sr.activity = 'active' AND sr.status_detail = 'pending' AND sr.status_detail != 'expired'
  AND NOW() - sr.updated_at > INTERVAL '36 hour'
RETURNING
  sr.id,
  sr.listing_id,
  sr.status_detail,
  sr.updated_at,
  sr.requester_id,
  sr.provider_id,
  sr.token_reward,
  sl.title AS listing_title,
  ru.full_name AS requester_full_name;


-- name: GetProvidingeRequests :many
select sr.id,sl.title from service_request sr
join service_listing sl 
on sl.id = sr.listing_id
where sr.activity  = 'active' and sr.provider_id = $1;



-- name: GetAllUserTickets :many
select rr.*,sl.title from request_report rr
join service_request sr 
on sr.id = rr.request_id
join service_listing sl
on sl.id = sr.listing_id 
where reporter_id = $1;