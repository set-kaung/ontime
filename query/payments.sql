-- name: InsertPaymentHolding :execresult
INSERT INTO payments(service_request_id,payer_id,status,amount_tokens,created_at,updated_at)
SELECT
    $1,
    $2,
    'holding',sr.token_reward,NOW(),NOW()
FROM service_requests sr
WHERE sr.id = $1
RETURNING id;

-- name: GetPaymentHolding :one
SELECT * FROM payments
WHERE service_request_id = $1 AND payer_id = $2;

-- name: UpdatePaymentHolding :execresult
UPDATE payments
SET status = $1, updated_at = NOW()
WHERE service_request_id = $2;
