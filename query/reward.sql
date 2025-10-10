-- name: GetAllRewards :many
SELECT r.*,COUNT(cc.id) as available_amount FROM reward r
JOIN coupon_code cc
ON cc.reward_id = r.id
WHERE cc.is_claimed = FALSE
GROUP BY r.id;


-- name: GetAllUserRedeemdRewards :many
SELECT rr.id,rr.reward_id,rr.user_id,rr.redeemed_at,rr.cost as redeemed_cost,r.title,r.description,r.image_url,cc.coupon_code FROM redeemed_reward rr
JOIN reward r
ON r.id = rr.reward_id
JOIN coupon_code cc
ON cc.id = rr.coupon_code_id
WHERE rr.user_id = $1;


-- name: GetRewardByID :one
SELECT r.*,COUNT(cc.id) as available_amount FROM reward r
JOIN coupon_code cc
ON cc.reward_id = r.id
WHERE r.id = $1
GROUP BY r.id;


-- name: InsertRedeemedReward :execrows
INSERT INTO redeemed_reward (reward_id, user_id, redeemed_at, cost,coupon_code_id)
    SELECT
        $1,
        $2,
        NOW(),
        r.cost,
        $3
FROM reward r
JOIN "user" u ON u.id = $2
WHERE r.id = $1 AND u.token_balance >= r.cost;

-- name: UpdateCouponCodeStatus :execrows
UPDATE coupon_code
SET is_claimed = TRUE
WHERE id = $1;

-- name: GetAllCouponCodes :many
SELECT * FROM coupon_code
WHERE reward_id = $1;


-- name: GetRedeemedRewardByID :one
SELECT
  rr.id as redeemed_id,rr.reward_id,rr.redeemed_at,rr.user_id,rr.cost,
  r.title,r.description,cc.coupon_code,r.image_url,
  cc.coupon_code
FROM redeemed_reward rr
JOIN reward r
ON r.id = rr.reward_id
JOIN coupon_code cc
ON cc.id = rr.coupon_code_id
WHERE rr.id = $1;
