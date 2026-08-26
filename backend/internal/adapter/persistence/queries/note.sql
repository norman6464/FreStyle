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

-- name: InsertNote :one
-- 学習メモを 1 件作成する。created_at / updated_at は DB 既定値が無いため呼び出し側が
-- 値を渡す（GORM autoCreateTime/autoUpdateTime 相当。ゼロなら呼び出し側で now() を入れる）。
-- RETURNING で id / created_at / updated_at を書き戻す。
INSERT INTO notes (user_id, title, content, is_public, is_pinned, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at, updated_at;

-- name: UpdateNote :one
-- メモ本文を更新する（所有者検証は usecase 側で済ませてから呼ぶ）。updated_at は now() へ
-- 進める（GORM autoUpdateTime 相当）。created_at は触らない。RETURNING で updated_at を書き戻す。
UPDATE notes SET
  title     = $2,
  content   = $3,
  is_public = $4,
  is_pinned = $5,
  updated_at = now()
WHERE id = $1
RETURNING updated_at;

-- name: DeleteNote :exec
-- メモを削除する。user_id で絞り、他人のメモを消せないようにする（notes に論理削除列は無い）。
DELETE FROM notes
WHERE id = $1 AND user_id = $2;
