-- name: GetAdsWatched :one
select count(id) from ads_watching_history
WHERE user_id = $1 AND date_time > (NOW() - INTERVAL '24 hour');

-- name: InsertAdsHistory :execresult
INSERT INTO ads_watching_history (user_id,date_time)
VALUES ($1,NOW());
