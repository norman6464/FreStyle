-- users の読み出し（FRESTYLE-311 移行期間中）。
-- role_name は旧カラム u.role を正とし、空のときだけ roles.name に落とす
-- （移行期間中は旧コードも書くため旧カラムが常に最新。role_id は起動時バックフィルで追随する）。
-- PR3（旧カラム DROP）で r.name 基準へ一斉に切り替える。

-- name: GetUserByCognitoSub :one
-- OIDC subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
-- 正は user_oidc_identities。旧コードが identities 未作成のまま挿入した行にも
-- 旧カラム users.cognito_sub のフォールバックで到達できるようにする。
SELECT u.*, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
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
SELECT u.*, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.*, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE COALESCE(NULLIF(u.role, ''), r.name, '') = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByCompanyID :many
-- 会社単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.*, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: GetActiveUserByEmail :one
-- email で有効ユーザーを 1 件引く（論理削除・無効化は除外）。ローカルのパスワードログイン用。
-- email は uq_users_email_active（deleted_at IS NULL AND email <> ''）でアクティブ行に対して一意。
SELECT u.*, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.email = $1 AND u.email <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの cognito provider の OIDC subject を引く。ローカルのパスワードログインが
-- 発行するトークンの sub に使う（無ければ呼び出し側が生成して EnsureOidcIdentity する）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';
