-- users の読み出し（FRESTYLE-311 正規化完了）。旧カラム users.role / users.cognito_sub は
-- 撤去済み（migrations/0021）のため参照しない。
-- role_name は roles マスタを JOIN して解決する（正は users.role_id → roles.name）。
-- OIDC subject の突き合わせは user_oidc_identities のみで行う。
--
-- password_hash はローカルのパスワードログイン専用の GetActiveUserByEmail だけが取得する
-- （一覧・認証解決の経路で bcrypt ハッシュをアプリメモリに載せない）。

-- name: GetUserByCognitoSub :one
-- OIDC subject で 1 ユーザーを引く（論理削除は除外）。認証時の user 解決に使う。
-- 正は user_oidc_identities（provider='cognito' の subject）。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.is_platform_admin, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
  AND u.id IN (
    SELECT oi.user_id FROM user_oidc_identities oi
    WHERE oi.provider = 'cognito' AND oi.subject = $1
  );

-- name: GetUserByID :one
-- 内部 ID で 1 ユーザーを引く（論理削除は除外）。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.is_platform_admin, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.is_platform_admin, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE r.name = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByCompanyID :many
-- 会社単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.is_platform_admin, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.company_id = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListActiveUsersByEmail :many
-- email で有効ユーザーを引く（論理削除・無効化は除外）。ローカルのパスワードログイン専用で、
-- ハッシュを含む唯一のクエリ。email は uq_users_email_active（lower(btrim(email, ...)) /
-- deleted_at IS NULL AND btrim(email, ...) <> ''）でアクティブ行に対して一意だが、既存データの
-- 重複で index 未作成のまま起動している環境では複数行になり得るため :many で受け、
-- 呼び出し側が曖昧さを拒否する。
-- 突き合わせは domain.NormalizeEmail と同じ正規形 lower(btrim(email, E'\t\n\x0B\f\r ')) で行う
-- （索引・述語と同じ式なのでそのまま部分索引が使われ、正規化される前に入った大文字混じり・
-- 前後空白付きの既存行も同じアドレスとして 1 つに解決される）。引数側も同じ式で畳むので、
-- ログインフォームの生入力をそのまま渡してよい（引数は ::text を明示する。btrim には bytea
-- 版もあり、キャストが無いと sqlc が引数を []byte と推論してしまう）。
SELECT u.id, u.email, u.name, u.company_id, u.role_id, u.ai_chat_enabled, u.is_active, u.is_platform_admin, u.created_at, u.updated_at, u.deleted_at, u.password_hash, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE lower(btrim(u.email, E'\t\n\x0B\f\r ')) = lower(btrim(sqlc.arg(email)::text, E'\t\n\x0B\f\r '))
  AND btrim(u.email, E'\t\n\x0B\f\r ') <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの cognito provider の OIDC subject を引く。ローカルのパスワードログインが
-- 発行するトークンの sub に使う（無ければ呼び出し側が生成して EnsureOidcIdentity する）。
-- (user_id, provider) は uq_user_oidc_user_provider で一意（最大 1 行）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';
