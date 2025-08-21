-- name: InsertNotification :execresult
INSERT INTO notifications (message,recipient_user_id,action_user_id,is_read,event_id)
VALUES ($1,$2,$3,false,$4);


-- name: GetNotifications :many
SELECT * FROM notifications n
JOIN notification_events ne ON ne.id = n.event_id
WHERE recipient_user_id = $1;


-- name: InsertEvent :one
INSERT INTO notification_events (target_id,"type",created_at)
VALUES ($1,$2,NOW())
RETURNING id;

-- name: GetUnreadNotificationsCount :one
SELECT COUNT(n.id) FROM notifications n
WHERE n.recipient_user_id = $1 AND n.is_read = false;

-- name: SetUserNotificationsRead :execresult
UPDATE notifications
SET is_read = true
WHERE id = $1 AND recipient_user_id = $2 AND is_read = false;
