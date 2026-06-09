-- name: GetUserByCognitoSub :one
-- Cognito subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
SELECT * FROM users
WHERE cognito_sub = $1 AND deleted_at IS NULL;

-- name: GetUserByID :one
-- 内部 ID で 1 ユーザーを引く（論理削除は除外）。
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT * FROM users
WHERE role = $1 AND deleted_at IS NULL
ORDER BY id ASC;
