-- users の読み出し（FRESTYLE-311 移行期間中）。
-- role_name は旧カラム u.role を正とし、空のときだけ roles.name に落とす
-- （移行期間中は旧コードも書くため旧カラムが常に最新。role_id は起動時バックフィルで追随する）。
-- PR3（旧カラム DROP）で r.name 基準へ一斉に切り替える。
--
-- password_hash はローカルのパスワードログイン専用の GetActiveUserByEmail だけが取得する
-- （一覧・認証解決の経路で bcrypt ハッシュをアプリメモリに載せない）。

-- name: GetUserByCognitoSub :one
-- OIDC subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
-- 正は user_oidc_identities。旧コードが identities 未作成のまま挿入した行にも
-- 旧カラム users.cognito_sub のフォールバックで到達できるようにする。
SELECT u.id, u.cognito_sub, u.email, u.name, u.company_id, u.role, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
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
SELECT u.id, u.cognito_sub, u.email, u.name, u.company_id, u.role, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.id, u.cognito_sub, u.email, u.name, u.company_id, u.role, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE COALESCE(NULLIF(u.role, ''), r.name, '') = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByCompanyID :many
-- 会社単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.id, u.cognito_sub, u.email, u.name, u.company_id, u.role, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListActiveUsersByEmail :many
-- email で有効ユーザーを引く（論理削除・無効化は除外）。ローカルのパスワードログイン専用で、
-- ハッシュを含む唯一のクエリ。email は uq_users_email_active（deleted_at IS NULL AND
-- email <> ''）でアクティブ行に対して一意だが、既存データの重複で index 未作成のまま
-- 起動している環境では複数行になり得るため :many で受け、呼び出し側が曖昧さを拒否する。
SELECT u.id, u.cognito_sub, u.email, u.name, u.company_id, u.role, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, u.password_hash, COALESCE(NULLIF(u.role, ''), r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.email = $1 AND u.email <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの cognito provider の OIDC subject を引く。ローカルのパスワードログインが
-- 発行するトークンの sub に使う（無ければ呼び出し側が生成して EnsureOidcIdentity する）。
-- (user_id, provider) は uq_user_oidc_user_provider で一意（最大 1 行）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';
