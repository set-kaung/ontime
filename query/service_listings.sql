-- name: GetUserListings :many
SELECT * FROM service_listings
WHERE posted_by = $1 AND status = 'active';

-- name: InsertListing :one
INSERT INTO service_listings (title,"description",token_reward,posted_by,category,image_url,posted_at,status,session_duration,contact_method)
VALUES ($1, $2, $3, $4,$5,$6, NOW(),'active',$7,$8)
RETURNING id;

-- name: DeleteListing :execresult
UPDATE service_listings
SET status = 'inactive'
WHERE id = $1 AND posted_by = $2;


-- name: GetListingByID :one
SELECT sl.*,u.id uid,u.full_name,sr.id as request_id,r.total_ratings,r.rating_count FROM service_listings sl
JOIN users u
ON u.id = sl.posted_by
LEFT JOIN service_requests sr ON sr.listing_id = sl.id AND sr.activity = 'active' AND sr.requester_id = $2
LEFT JOIN ratings r ON r.user_id = sl.posted_by
WHERE sl.id = $1 and sl.status = 'active';

-- name: GetAllListings :many
SELECT sl.*,u.id uid,u.full_name FROM service_listings sl
JOIN users u
ON u.id = sl.posted_by
WHERE posted_by != $1 AND sl.status = 'active';

-- name: UpdateListing :execrows
UPDATE service_listings
SET title = $1, description = $2, token_reward = $3, category=$4, image_url = $5, session_duration = $6, contact_method = $7
WHERE id = $8 AND posted_by = $9;
