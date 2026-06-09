-- name: ListNotificationsByUserID :many
-- 自分の通知を新しい順で返す。
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CountUnreadNotifications :one
-- 未読通知数（バッジ表示用）。
SELECT count(*) FROM notifications
WHERE user_id = $1 AND is_read = false;
