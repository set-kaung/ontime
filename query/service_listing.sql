-- name: GetUserListings :many
SELECT sl.*,w.id as warning_id,w.severity,w.created_at as warning_created_at,w.comment as warning_comment,w.reason as warning_reason FROM service_listing sl
LEFT JOIN warning w
ON w.listing_id = sl.id
WHERE sl.posted_by = $1 AND sl.status = 'active';

-- name: InsertListing :one
INSERT INTO service_listing (title,"description",token_reward,posted_by,category,image_url,posted_at,status,session_duration,contact_method)
VALUES ($1, $2, $3, $4,$5,$6, NOW(),'active',$7,$8)
RETURNING id;

-- name: DeleteListing :execresult
UPDATE service_listing
SET status = 'inactive'
WHERE id = $1 AND posted_by = $2;


-- name: GetListingByID :one
SELECT sl.*,u.id uid,u.full_name,sr.id as request_id,r.total_ratings,r.rating_count,w.id as warning_id,w.severity,w.created_at as warning_created_at,w.comment as warning_comment,w.reason as warning_reason FROM service_listing sl
JOIN "user" u
ON u.id = sl.posted_by
LEFT JOIN service_request sr ON sr.listing_id = sl.id AND sr.activity = 'active' AND sr.requester_id = $2
LEFT JOIN rating r ON r.user_id = sl.posted_by
LEFT JOIN warning w
ON w.listing_id = w.id
WHERE sl.id = $1 and sl.status = 'active';

-- name: GetAllListings :many
with listing_rating as (
select sl.id as listing_id ,sum(r.rating ) as total_rating,count(r.id )as rating_count from service_listing sl
join service_request sr
on sr.listing_id = sl.id
join review r
on r.request_id  = sr.id
group by sl.id)
SELECT sl.*,u.id uid,u.full_name,coalesce(lr.rating_count,0) as total_rating_count ,coalesce(lr.total_rating,0) as total_ratings FROM service_listing sl
JOIN "user" u
ON u.id = sl.posted_by
LEFT JOIN listing_rating lr
ON lr.listing_id = sl.id
WHERE sl.posted_by != $1 AND sl.status = 'active';

-- name: UpdateListing :execrows
UPDATE service_listing
SET title = $1, description = $2, token_reward = $3, category=$4, image_url = $5, session_duration = $6, contact_method = $7
WHERE id = $8 AND posted_by = $9;


-- name: GetPartialListingsByUserID :many
with listing_rating as (
select sl.id as listing_id ,sum(r.rating ) as total_rating,count(r.id )as rating_count from service_listing sl
join service_request sr
on sr.listing_id = sl.id
join review r
on r.request_id  = sr.id
group by sl.id)
select sl.id,sl.title,sl.token_reward ,sl.posted_at,sl.category,sl.image_url, coalesce(lr.rating_count,0) ,coalesce(lr.total_rating,0)  from service_listing sl
left join listing_rating lr
on lr.listing_id = sl.id
where sl.posted_by = $1 and status = 'active';
