-- name: InsertServiceRequestReview :one
INSERT INTO reviews (request_id,reviewer_id,reviewee_id,rating,comment,date_time)
VALUES ($1,$2,(SELECT provider_id from service_requests WHERE id = $1),$3,$4,NOW())
RETURNING id, reviewee_id;

-- name: UpdateUserRating :one
UPDATE ratings
SET total_ratings = total_ratings + $1,
    rating_count  = rating_count + 1
WHERE user_id = (
    SELECT provider_id FROM service_requests
    WHERE id = sqlc.arg(request_id)
)
RETURNING *;


-- name: InsertNewUserRating :exec
INSERT INTO ratings (user_id,total_ratings, rating_count)
VALUES ($1,0,0);


-- name: GetReviewByID :one
SELECT * FROM reviews
WHERE id = $1;


-- name: GetReviewByRequestID :one
SELECT * FROM reviews
WHERE request_id = $1;
