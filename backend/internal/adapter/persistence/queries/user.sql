-- name: GetUserByCognitoSub :one
SELECT u.id, u.email, u.name, u.workspace_id, u.is_active, u.created_at, u.updated_at, u.deleted_at
FROM users u
WHERE u.deleted_at IS NULL
  AND u.id IN (
    SELECT oi.user_id FROM user_oidc_identities oi
    WHERE oi.provider = 'cognito' AND oi.subject = $1
  );

-- name: GetUserByID :one
-- 内部 ID で 1 ユーザーを引く（論理削除は除外）。
SELECT u.id, u.email, u.name, u.workspace_id, u.is_active, u.created_at, u.updated_at, u.deleted_at
FROM users u
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByWorkspaceID :many
-- ワークスペース単位のユーザー一覧（論理削除は除外）。
SELECT u.id, u.email, u.name, u.workspace_id, u.is_active, u.created_at, u.updated_at, u.deleted_at
FROM users u
WHERE u.workspace_id = $1 AND u.deleted_at IS NULL
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
SELECT u.id, u.email, u.name, u.workspace_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, u.password_hash
FROM users u
WHERE lower(btrim(u.email, E'\t\n\x0B\f\r ')) = lower(btrim(sqlc.arg(email)::text, E'\t\n\x0B\f\r '))
  AND btrim(u.email, E'\t\n\x0B\f\r ') <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの OIDC subject を引く。
-- (user_id, provider) は uq_user_oidc_user_provider で一意（最大 1 行）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';

-- name: InsertUser :one
-- ユーザーを 1 件作る（id は採番シーケンスに任せる）。created_at / updated_at は DB 既定値が
-- 無いため呼び出し側が値を渡す。is_active は常に true（作成直後のアカウントは有効。無効化は
-- UpdateUserActive の仕事）。RETURNING で id / created_at / updated_at を書き戻す。
--
-- workspace_id は呼び出し側が解決した値をそのまま書く（companies へのサブクエリ参照はしない）。
INSERT INTO users (
  email, password_hash, name, workspace_id,
  is_active, created_at, updated_at, deleted_at
)
VALUES (
  $1, $2, $3, $4, true, $5, $6, $7
)
RETURNING id, created_at, updated_at;

-- name: InsertUserWithID :one
-- id を呼び出し側が決める場合の InsertUser。列と既定の扱いは InsertUser と同じにすること
-- （片方だけ列を足すと、id を指定する経路だけ値が入らない）。
INSERT INTO users (
  id, email, password_hash, name, workspace_id,
  is_active, created_at, updated_at, deleted_at
)
VALUES (
  $1, $2, $3, $4, $5, true, $6, $7, $8
)
RETURNING id, created_at, updated_at;

-- name: InsertOidcIdentityIfAbsent :execrows
-- OIDC identity を冪等に挿入する。既に同じ (provider, subject) があれば 0 行で、
-- 呼び出し側が持ち主を確かめる。(user_id, provider) の一意制約違反はそのままエラーになる。
INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
ON CONFLICT (provider, subject) DO NOTHING;

-- name: GetOidcIdentityOwner :one
-- (provider, subject) を持っているユーザーの id。挿入されなかったときの持ち主判定に使う。
SELECT user_id FROM user_oidc_identities
WHERE provider = $1 AND subject = $2;

-- name: DeleteOidcIdentitiesByUserID :exec
-- ユーザーの OIDC identity をすべて消し、subject の占有を解く（同じアカウントの再招待を可能にする）。
DELETE FROM user_oidc_identities WHERE user_id = $1;

-- name: UpdateUserActive :execrows
-- アカウントの有効/無効を更新する。0 件なら対象が存在しない（呼び出し側が not-found にする）。
UPDATE users SET is_active = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserName :execrows
-- 氏名だけを更新する。0 件なら対象の user が存在しない（呼び出し側が not-found にする）。
UPDATE users SET name = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserWorkspaceID :execrows
-- 所属ワークスペースを付け替える。呼び出し側が既に解決した workspace_id をそのまま書く。
-- ワークスペースが無いユーザーもあり得るため nullable。
-- 0 件なら対象の user が存在しない（呼び出し側が not-found にする）。
UPDATE users SET
  workspace_id = sqlc.narg(workspace_id)
WHERE id = sqlc.arg(id);

-- name: SoftDeleteUser :execrows
-- ユーザーを論理削除する。既に削除済み / 存在しない場合は 0 件（呼び出し側が not-found にする）。
UPDATE users SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
