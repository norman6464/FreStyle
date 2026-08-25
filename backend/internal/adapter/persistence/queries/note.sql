-- name: ListNotesByUserID :many
-- 自分の学習メモ一覧を更新日時の新しい順で返す。
-- updated_at は一意でないため、同時刻の順序を固定する id DESC をタイブレークに付ける。
SELECT * FROM notes
WHERE user_id = $1
ORDER BY updated_at DESC, id DESC;

-- name: GetNoteByID :one
-- 内部 ID で 1 件取得（所有者検証は usecase 側で user_id を突き合わせる）。
SELECT * FROM notes
WHERE id = $1;
