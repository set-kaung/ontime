-- name: InsertServiceRequestReview :one
INSERT INTO reviews (request_id,reviewer_id,reviewee_id,rating,comment,date_time)
VALUES ($1,$2,(SELECT provider_id from service_requests WHERE id = $1),$3,$4,NOW())
RETURNING id, reviewee_id;

-- name: UpdateUserRating :exec
UPDATE ratings
SET total_ratings = total_ratings + $1, rating_count = rating_count + 1
WHERE user_id = $2;


-- name: InsertNewUserRating :exec
INSERT INTO ratings (user_id,total_ratings, rating_count)
VALUES ($1,0,0);


-- name: GetReviewByID :one
SELECT * FROM reviews
WHERE id = $1;
