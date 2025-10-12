-- name: InsertPaymentHolding :one
INSERT INTO payment(service_request_id,payer_id,status,amount_tokens,created_at,updated_at)
SELECT
    $1,
    $2,
    'holding',sr.token_reward,NOW(),NOW()
FROM service_request sr
WHERE sr.id = $1
RETURNING id;

-- name: GetPaymentHolding :one
SELECT * FROM payment
WHERE service_request_id = $1 AND payer_id = $2;

-- name: UpdatePaymentHolding :execresult
UPDATE payment
SET status = $1, updated_at = NOW()
WHERE service_request_id = $2;

-- name: GetRequestPayment :one
SELECT * FROM payment
WHERE service_request_id = $1;
