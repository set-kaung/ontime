-- name: GetUserByID :one
SELECT
    u.*,
    COALESCE(sp.requested_count, 0) AS services_received,
    COALESCE(sp.provided_count, 0) AS services_provided,
    r.total_ratings,
    r.rating_count
FROM "user" u
LEFT JOIN (
    SELECT
      user_id,
      COUNT(*) FILTER (WHERE role = 'requester') AS requested_count,
      COUNT(*) FILTER (WHERE role = 'provider') AS provided_count
    FROM (
      SELECT requester_id AS user_id, 'requester' AS role
      FROM service_request
      UNION ALL
      SELECT provider_id AS user_id, 'provider' AS role
      FROM service_request
    ) combined
    GROUP BY user_id
) sp ON u.id = sp.user_id
LEFT JOIN rating r
ON r.user_id = u.id
WHERE u.id = $1;

-- name: InsertUser :one
INSERT INTO "user" (
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
    is_email_signedup,
    is_paid
)
VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
  $11, NOW(), $12,false
)
RETURNING id;





-- name: UpdateUser :execresult
UPDATE "user"
SET full_name = $1, phone = $2, address_line_1 = $3, address_line_2 = $4, city = $5, state_province = $6, zip_postal_code = $7, country = $8
WHERE id = $9;

-- name: DeleteUser :execresult
DELETE FROM "user" where id = $1;

-- name: GetUserTokenBalance :one
SELECT token_balance FROM "user"
WHERE id = $1;

-- name: AddTokens :one
UPDATE "user"
SET token_balance = token_balance + $1
WHERE id = $2
RETURNING token_balance;

-- name: MarkSignupPaidAndAward :one
UPDATE "user"
SET
  is_paid = true,
  token_balance = token_balance + $2
WHERE id = $1
  AND is_paid = false
RETURNING token_balance;


-- name: UpdateAboutMe :exec
UPDATE "user"
SET about_me = $1
WHERE id = $2;


-- name: DeductTokens :execrows
UPDATE "user"
SET token_balance = token_balance - s.token_reward
FROM service_listing s
WHERE "user".id = sqlc.arg(user_id)
  AND s.id = sqlc.arg(listing_id)
  AND "user".token_balance >= s.token_reward;

-- name: DeductRewardTokensFromUser :exec
UPDATE "user"
SET token_balance = token_balance - r.cost
FROM reward r
WHERE "user".id = sqlc.arg(user_id) AND r.id = sqlc.arg(reward_id) AND "user".token_balance >= r.cost;

-- name: UpdateUserFullNmae :execrows
UPDATE "user"
SET full_name = $1
WHERE id = $2;


-- name: GetUserFullNameByID :one
SELECT full_name FROM "user"
WHERE id = $1;


