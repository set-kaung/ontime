-- name: InsertTransaction :exec
INSERT INTO "transaction" (user_id,type,amount,created_at)
SELECT
$1,$2,p.amount_tokens,NOW()
FROM payment p
WHERE p.id = sqlc.arg(payment_id);
