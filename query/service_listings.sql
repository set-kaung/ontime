-- name: GetUserListings :many
SELECT * FROM service_listings
WHERE posted_by = $1;

-- name: InsertListing :one
INSERT INTO service_listings (title,"description",token_reward,posted_by,category,image_url,posted_at)
VALUES ($1, $2, $3, $4,$5,$6, NOW())
RETURNING id;

-- name: DeleteListing :execresult
DELETE FROM service_listings
WHERE id = $1 AND posted_by = $2;


-- name: GetListingByID :one
SELECT sl.id,sl.title,sl.description,sl.token_reward,sl.posted_at,sl.category,sl.image_url,u.id uid,u.full_name,sr.id as request_id,r.total_ratings,r.rating_count FROM service_listings sl
JOIN users u
ON u.id = sl.posted_by
LEFT JOIN service_requests sr ON sr.listing_id = sl.id AND sr.activity = 'active' AND sr.requester_id = $2
LEFT JOIN ratings r ON r.user_id = sl.posted_by
WHERE sl.id = $1;

-- name: GetAllListings :many
SELECT sl.id,sl.title,sl.description,sl.token_reward,sl.posted_at,sl.category,sl.image_url,u.id uid,u.full_name FROM service_listings sl
JOIN users u
ON u.id = sl.posted_by
WHERE posted_by != $1;

-- name: UpdateListing :execrows
UPDATE service_listings
SET title = $1, description = $2, token_reward = $3, category=$4, image_url = $5
WHERE id = $6 AND posted_by = $7;

-- name: GetListingReviews :many
select r.*,sr.listing_id,reviewer_user.full_name as reviewer_full_name,reviewee_user.full_name as reviewee_full_name from reviews r
JOIN service_requests sr
ON sr.id = r.request_id
JOIN users reviewer_user
ON reviewer_user.id = r.reviewer_id
JOIN users reviewee_user
ON reviwee_user.id = r.reviewee_id
WHERE listing_id = $1;