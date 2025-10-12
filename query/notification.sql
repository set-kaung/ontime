-- name: InsertNotification :execresult
INSERT INTO notification (message,recipient_user_id,action_user_id,is_read,event_id)
VALUES ($1,$2,$3,false,$4);


-- name: GetNotifications :many
SELECT * FROM notification n
JOIN "event" ne ON ne.id = n.event_id
WHERE recipient_user_id = $1;


-- name: InsertEvent :one
INSERT INTO "event" (target_id,"type",created_at,description)
VALUES ($1,$2,NOW(),$3)
RETURNING id;

-- name: GetUnreadNotificationCount :one
SELECT COUNT(n.id) FROM notification n
WHERE n.recipient_user_id = $1 AND n.is_read = false;

-- name: SetUserNotificationsRead :execresult
UPDATE notification
SET is_read = true
WHERE id = $1 AND recipient_user_id = $2 AND is_read = false;


-- name: SetAllNotificationsRead :exec
UPDATE notification
SET is_read = true
WHERE recipient_user_id = $1
  AND event_id IN (
    SELECT id FROM "event"
    WHERE created_at < $2
  );


-- name: GetAllEventOfARequest :many
SELECT e.*, n.action_user_id FROM "event" e
JOIN notification n
ON n.event_id = e.id
WHERE target_id = $1 AND type = 'request'
ORDER BY created_at DESC;
