-- name: GetAllRewards :many
SELECT id,title,description,cost,available_amount,image_url,created_date FROM rewards
WHERE available_amount > 0;


-- name: GetAllUserRedeemdRewards :many
SELECT rr.id,rr.reward_id,rr.user_id,rr.redeemed_at,rr.cost as redeemed_cost,r.title FROM redeemed_rewards rr
LEFT JOIN rewards r
ON r.id = rr.reward_id
WHERE rr.user_id = $1;


-- name: GetRewardByID :one
SELECT * FROM rewards
WHERE id = $1;


-- name: InsertRedeemedReward :one
WITH inserted AS (
    INSERT INTO redeemed_rewards (reward_id, user_id, redeemed_at, cost)
    SELECT
        $1,
        $2,
        NOW(),
        r.cost
    FROM rewards r
    JOIN users u ON u.id = $2
    WHERE r.id = $1 AND u.token_balance > 0
    RETURNING reward_id
)
SELECT r.coupon_code
FROM inserted i
JOIN rewards r ON r.id = i.reward_id;

-- name: DeductRewardAmount :execrows
UPDATE rewards
SET available_amount = available_amount - 1
WHERE id = $1 AND available_amount > 0;
