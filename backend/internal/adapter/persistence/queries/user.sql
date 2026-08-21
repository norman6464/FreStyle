-- users の読み出し（FRESTYLE-311 正規化・expand-contract の読み替えフェーズ）。
-- role_name は roles マスタを JOIN して解決する（正は users.role_id → roles.name）。旧 role 列は読まない。
-- cognito_sub は GetUserByCognitoSub の WHERE でフォールバックとしてのみ参照する（ローリングデプロイ中の
-- 孤児行救済。旧カラムの物理撤去とこのフォールバック除去は後続 PR）。旧カラムを SELECT には含めない。
--
-- password_hash はローカルのパスワードログイン専用の GetActiveUserByEmail だけが取得する
-- （一覧・認証解決の経路で bcrypt ハッシュをアプリメモリに載せない）。

-- name: GetUserByCognitoSub :one
-- OIDC subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
-- 正は user_oidc_identities（provider='cognito' の subject）だが、旧タスクが identity 未作成の
-- まま挿入した孤児行にも旧カラム users.cognito_sub のフォールバックで到達できるようにする
-- （ローリングデプロイ中の可用性保全。このフォールバックは旧カラム撤去の後続 PR で外す）。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
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
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE r.name = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByCompanyID :many
-- 会社単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListActiveUsersByEmail :many
-- email で有効ユーザーを引く（論理削除・無効化は除外）。ローカルのパスワードログイン専用で、
-- ハッシュを含む唯一のクエリ。email は uq_users_email_active（deleted_at IS NULL AND
-- email <> ''）でアクティブ行に対して一意だが、既存データの重複で index 未作成のまま
-- 起動している環境では複数行になり得るため :many で受け、呼び出し側が曖昧さを拒否する。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.created_at, u.updated_at, u.deleted_at, u.password_hash, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.email = $1 AND u.email <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの cognito provider の OIDC subject を引く。ローカルのパスワードログインが
-- 発行するトークンの sub に使う（無ければ呼び出し側が生成して EnsureOidcIdentity する）。
-- (user_id, provider) は uq_user_oidc_user_provider で一意（最大 1 行）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';
