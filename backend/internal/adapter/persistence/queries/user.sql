-- name: GetUserByCognitoSub :one
SELECT u.id, u.email, u.name, u.workspace_id, u.role_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.deleted_at IS NULL
  AND u.id IN (
    SELECT oi.user_id FROM user_oidc_identities oi
    WHERE oi.provider = 'cognito' AND oi.subject = $1
  );

-- name: GetUserByID :one
-- 内部 ID で 1 ユーザーを引く（論理削除は除外）。
SELECT u.id, u.email, u.name, u.workspace_id, u.role_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.id = $1 AND u.deleted_at IS NULL;

-- name: ListUsersByRole :many
-- role 名単位の一覧（論理削除は除外）。super_admin / company_admin の管理画面用。
SELECT u.id, u.email, u.name, u.workspace_id, u.role_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE r.name = $1 AND u.deleted_at IS NULL
ORDER BY u.id ASC;

-- name: ListUsersByWorkspaceID :many
-- ワークスペース単位の従業員一覧（論理削除は除外）。company_admin の従業員管理画面用。
SELECT u.id, u.email, u.name, u.workspace_id, u.role_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
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
SELECT u.id, u.email, u.name, u.workspace_id, u.role_id, u.is_active, u.created_at, u.updated_at, u.deleted_at, u.password_hash, COALESCE(r.name, '') AS role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE lower(btrim(u.email, E'\t\n\x0B\f\r ')) = lower(btrim(sqlc.arg(email)::text, E'\t\n\x0B\f\r '))
  AND btrim(u.email, E'\t\n\x0B\f\r ') <> '' AND u.deleted_at IS NULL AND u.is_active;

-- name: GetCognitoSubjectByUserID :one
-- ユーザーの OIDC subject を引く。
-- (user_id, provider) は uq_user_oidc_user_provider で一意（最大 1 行）。
SELECT subject FROM user_oidc_identities
WHERE user_id = $1 AND provider = 'cognito';

-- name: GetRoleIDByName :one
-- ロール名を roles.id に解決する。未知の名前は 0 件で返り、呼び出し側がエラーにする
-- （黙って別ロールへ倒さない）。
SELECT id FROM roles WHERE name = $1;

-- name: AcquireBootstrapSuperAdminLock :exec
-- 「最初の運営管理者を作る」経路を直列化するロックを取る。
--
-- トランザクションスコープのロック（pg_advisory_xact_lock）なので、必ず判定と INSERT と
-- 同じトランザクション（Queries.WithTx）で発行すること。別接続で取ると、ロックが取れた
-- 直後に解放され「0 人か確かめて作る」の間を守れない。pgbouncer（transaction pooler）が
-- 接続を貸し借りする本番でセッションロックが使えないのも同じ理由。
SELECT pg_advisory_xact_lock(sqlc.arg(lock_key)::bigint);

-- name: CountActiveSuperAdmins :one
-- 論理削除されていない運営管理者の人数。免除経路（招待なしの作成）が既に閉じているかの判定に使う。
SELECT count(*) FROM users u
JOIN roles r ON r.id = u.role_id
WHERE r.name = $1 AND u.deleted_at IS NULL;

-- name: InsertUser :one
-- ユーザーを 1 件作る（id は採番シーケンスに任せる）。created_at / updated_at は DB 既定値が
-- 無いため呼び出し側が値を渡す。is_active は常に true（作成直後のアカウントは有効。無効化は
-- UpdateUserActive の仕事）。RETURNING で id / created_at / updated_at を書き戻す。
--
-- workspace_id は呼び出し側が解決した値をそのまま書く（companies へのサブクエリ参照はしない）。
INSERT INTO users (
  email, password_hash, name, workspace_id, role_id,
  is_active, created_at, updated_at, deleted_at
)
VALUES (
  $1, $2, $3, $4, $5, true, $6, $7, $8
)
RETURNING id, created_at, updated_at;

-- name: InsertUserWithID :one
-- id を呼び出し側が決める場合の InsertUser。列と既定の扱いは InsertUser と同じにすること
-- （片方だけ列を足すと、id を指定する経路だけ値が入らない）。
INSERT INTO users (
  id, email, password_hash, name, workspace_id, role_id,
  is_active, created_at, updated_at, deleted_at
)
VALUES (
  $1, $2, $3, $4, $5, $6, true, $7, $8, $9
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

-- name: UpdateUserRoleID :execrows
-- 役割だけを更新する（誰がどの役割になれるかの判定は usecase 側の仕事）。
-- 0 件なら対象の user が存在しない（呼び出し側が not-found にする）。昇格が 1 行も
-- 当たっていないのに成功を返すと、権限が上がったつもりの利用者が生まれる。
UPDATE users SET role_id = $2, updated_at = now() WHERE id = $1;

-- name: UpdateUserWorkspaceID :execrows
-- 所属ワークスペースを付け替える（招待の受諾で呼ばれる）。呼び出し側（招待受諾）が
-- 招待行の workspace_id を既に持っているため、それをそのまま書く。ワークスペースが
-- 無い会社もあり得るため nullable。
-- 0 件なら対象の user が存在しない（呼び出し側が not-found にする）。
UPDATE users SET
  workspace_id = sqlc.narg(workspace_id)
WHERE id = sqlc.arg(id);

-- name: SoftDeleteUser :execrows
-- ユーザーを論理削除する。既に削除済み / 存在しない場合は 0 件（呼び出し側が not-found にする）。
UPDATE users SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
