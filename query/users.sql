-- name: GetUserByID :one
SELECT
    u.*,
    COALESCE(sp.requested_count, 0) AS services_received,
    COALESCE(sp.provided_count, 0) AS services_provided
FROM users u
LEFT JOIN (
    SELECT
      user_id,
      COUNT(*) FILTER (WHERE role = 'requester') AS requested_count,
      COUNT(*) FILTER (WHERE role = 'provider') AS provided_count
    FROM (
      SELECT requester_id AS user_id, 'requester' AS role
      FROM service_requests
      UNION ALL
      SELECT provider_id AS user_id, 'provider' AS role
      FROM service_requests
    ) combined
    GROUP BY user_id
) sp ON u.id = sp.user_id
WHERE u.id = $1;

-- name: InsertUser :one
INSERT INTO users (
    id,
    full_name,
    phone,
    token_balance,
    status,
    address_line_1,
    address_line_2,
    city,
    state_province,
    zip_postal_code,
    country,
    joined_at,
    email
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, NOW(), $12
)
RETURNING id;



-- name: UpdateUser :execresult
UPDATE users
SET full_name = $1, phone = $2, address_line_1 = $3, address_line_2 = $4, city = $5, state_province = $6, zip_postal_code = $7, country = $8
WHERE id = $9;

-- name: DeleteUser :execresult
DELETE FROM users where id = $1;

-- name: GetUserTokenBalance :one
SELECT token_balance FROM users
WHERE id = $1;

-- name: AddTokens :exec
UPDATE users
SET token_balance = token_balance + $1
WHERE id = $2;



-- name: DeductTokens :execrows
UPDATE users
SET token_balance = token_balance - s.token_reward
FROM service_listings s
WHERE users.id = sqlc.arg(user_id)
  AND s.id = sqlc.arg(listing_id)
  AND users.token_balance >= s.token_reward;



-- name: UpdateUserFullNmae :execrows
UPDATE users
SET full_name = $1
WHERE id = $2;
