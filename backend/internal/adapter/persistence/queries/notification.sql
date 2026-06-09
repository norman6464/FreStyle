-- name: ListNotificationsByUserID :many
-- 自分の通知を新しい順で返す。同時刻の順序を安定させるため id DESC をタイブレークに付ける。
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: CountUnreadNotifications :one
-- 未読通知数（バッジ表示用）。
SELECT count(*) FROM notifications
WHERE user_id = $1 AND is_read = false;
