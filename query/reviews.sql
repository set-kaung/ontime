-- name: InsertServiceRequestReview :one
INSERT INTO review (request_id,reviewer_id,reviewee_id,rating,comment,date_time)
VALUES ($1,$2,(SELECT provider_id from service_request WHERE id = $1),$3,$4,NOW())
RETURNING id, reviewee_id;

-- name: UpdateUserRating :one
UPDATE rating
SET total_ratings = total_ratings + $1,
    rating_count  = rating_count + 1
WHERE user_id = (
    SELECT provider_id FROM service_request
    WHERE id = sqlc.arg(request_id)
)
RETURNING *;


-- name: InsertNewUserRating :exec
INSERT INTO rating (user_id,total_ratings, rating_count)
VALUES ($1,0,0);


-- name: GetReviewByID :one
SELECT * FROM review
WHERE id = $1;


-- name: GetReviewByRequestID :one
SELECT r.*,
    reviewer.full_name AS reviewer_full_name,
    reviewee.full_name AS reviewee_full_name
FROM review r
JOIN "user" AS reviewer
  ON reviewer.id = r.reviewer_id
JOIN "user" AS reviewee
  ON reviewee.id = r.reviewee_id
WHERE r.request_id = $1;

-- name: GetListingreview :many
SELECT r.*,
       sr.listing_id,
       reviewer.full_name AS reviewer_full_name,
       reviewee.full_name AS reviewee_full_name
FROM review r
JOIN service_request sr
  ON sr.id = r.request_id
JOIN "user" AS reviewer
  ON reviewer.id = r.reviewer_id
JOIN "user" AS reviewee
  ON reviewee.id = r.reviewee_id
WHERE sr.listing_id = $1;
