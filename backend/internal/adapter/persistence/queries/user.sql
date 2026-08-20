-- users の読み出しは正規化テーブル（FRESTYLE-311）を正とする。
-- role_name は roles.name。移行期間中（旧コードが書いた直後の行など）に備えて
-- 旧カラム users.role へ COALESCE でフォールバックする（PR3 の旧カラム撤去で解消）。

-- name: GetUserByCognitoSub :one
-- OIDC subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
-- 正は user_oidc_identities。旧コードが identities 未作成のまま挿入した行にも
-- 旧カラム users.cognito_sub のフォールバックで到達できるようにする。
SELECT u.*, COALESCE(r.name, u.role) AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
  AND (
    u.id IN (
      SELECT oi.user_id FROM user_oidc_identities oi
      WHERE oi.provider = 'cognito' AND oi.subject = $1
    )
    OR u.cognito_sub = $1
  );

-- name: GetUserByID :one
-- 内部 ID で 1 ユーザーを引く（論理削除は除外）。
SELECT u.*, COALESCE(r.name, u.role) AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.*, COALESCE(r.name, u.role) AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE COALESCE(r.name, u.role) = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByCompanyID :many
-- 会社単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.*, COALESCE(r.name, u.role) AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;
