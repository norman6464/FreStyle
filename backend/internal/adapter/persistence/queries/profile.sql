-- name: GetProfileByUserID :one
-- user_id でプロフィールを 1 件取得（無ければ usecase 側で空表示にフォールバック）。
SELECT * FROM profiles
WHERE user_id = $1;
