-- name: GetActiveUserServiceRequests :many
SELECT
    sr.*,
    requester.full_name AS requester_name,
    provider.full_name  AS provider_name,
    l.title
FROM
    service_requests sr
JOIN users requester ON sr.requester_id = requester.id
JOIN users provider  ON sr.provider_id = provider.id
JOIN service_listings l on sr.listing_id  = l.id
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

  src.requester_completed,
  src.provider_completed

FROM service_requests sr
JOIN service_listings sl ON sr.listing_id = sl.id
JOIN users ru ON sr.requester_id = ru.id
JOIN users pu ON sr.provider_id = pu.id
JOIN service_request_completion src ON sr.id = src.request_id

WHERE sr.id = $1;

-- name: InsertPendingServiceRequest :one
INSERT INTO service_requests (listing_id,requester_id,provider_id,status_detail,activity,created_at,updated_at,token_reward)
SELECT
    $1,
    $2,
    sl.posted_by,
    'pending', 'active', NOW(),NOW(),sl.token_reward
FROM service_listings sl
WHERE sl.id = $1 AND sl.posted_by != $2
RETURNING id;


-- name: UpdateServiceRequest :one
UPDATE service_requests
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
    service_requests sr
JOIN users requester ON sr.requester_id = requester.id
JOIN users provider  ON sr.provider_id = provider.id
JOIN service_listings l on sr.listing_id  = l.id
WHERE
    (sr.provider_id = sqlc.arg(user_id) OR sr.requester_id = sqlc.arg(user_id));
