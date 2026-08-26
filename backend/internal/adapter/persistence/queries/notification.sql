-- name: ListNotificationsByUserID :many
-- 自分の通知を新しい順で返す。同時刻の順序を安定させるため id DESC をタイブレークに付ける。
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC, id DESC;

-- name: CountUnreadNotifications :one
-- 未読通知数（バッジ表示用）。
SELECT count(*) FROM notifications
WHERE user_id = $1 AND is_read = false;

-- name: InsertNotification :one
-- 通知を 1 件作成する。created_at は DB 既定値が無いため呼び出し側が値を渡す
-- （GORM autoCreateTime 相当。ゼロなら呼び出し側で now() を入れる）。RETURNING で
-- id / created_at を書き戻す。
INSERT INTO notifications (user_id, type, title, body, is_read, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at;

-- name: CreateNotifications :exec
-- 複数の通知を 1 回の INSERT でまとめて作成する（宛先が増えても往復を増やさない）。
-- database/sql モードで配列を渡すと lib/pq 依存が増えるため、items は 1 個の json 配列で
-- 渡し、json_to_recordset で行へ展開する。created_at は DB 既定値が無いので now() を明示する。
INSERT INTO notifications (user_id, type, title, body, is_read, created_at)
SELECT x.user_id, x.type, x.title, x.body, x.is_read, now()
FROM json_to_recordset(sqlc.arg(items)::json)
  AS x(user_id bigint, type text, title text, body text, is_read boolean);

-- name: MarkNotificationRead :exec
-- 単一通知を既読化する。WHERE で user_id を絞り、他人の通知を既読化できないようにする。
UPDATE notifications SET is_read = true
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
-- 対象 user の未読通知をすべて既読化する。
UPDATE notifications SET is_read = true
WHERE user_id = $1 AND is_read = false;
