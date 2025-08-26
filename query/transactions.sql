-- name: InsertTransaction :exec
INSERT INTO transactions (user_id,type,amount,created_at)
SELECT
$1,$2,p.amount_tokens,NOW()
FROM payments p
WHERE p.id = sqlc.arg(payment_id);
