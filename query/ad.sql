-- name: GetAdsWatched :one
SELECT count(id) FROM ads_watching_history
WHERE user_id = $1 AND date_time > (NOW() - INTERVAL '24 hour');

-- name: InsertAdsHistory :execresult
INSERT INTO ads_watching_history (user_id,date_time)
VALUES ($1,NOW());


-- name: GetAdsHistory :many
SELECT * FROM ads_watching_history
WHERE user_id = $1;
