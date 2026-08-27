-- name: ListCompanies :many
-- 企業一覧（名前昇順）。name に一意制約は無いので同名企業の順序を固定する id ASC を付ける。
SELECT * FROM companies
ORDER BY name ASC, id ASC;

-- name: GetCompanyByID :one
-- ID で企業を 1 件取得。
SELECT * FROM companies
WHERE id = $1;

-- name: UpdateCompanyActive :execrows
-- 会社アカウントの有効/無効を更新する。0 件なら対象が存在しない（呼び出し側が not-found にする）。
-- 判定は companies 側の件数で行う。写し先（workspaces）の件数を見ると、まだ紐付いていない
-- 会社が not-found に化ける。
UPDATE companies SET is_active = $2, updated_at = now() WHERE id = $1;

-- name: MirrorCompanySettingsToWorkspace :exec
-- 会社のテナント設定を、対応する workspaces 行へ写す。
-- is_active は最終的に workspaces の列になる。今は companies が正本で、workspaces 側は
-- その写し。設定を書く経路が増えても写し忘れないよう、companies から写すこの 1 か所に
-- 集約する。まだワークスペースに紐付いていない会社（workspace_id IS NULL）は 0 件更新で、
-- 写す先が無い。
UPDATE workspaces w
SET is_active = c.is_active,
    updated_at = now()
FROM companies c
WHERE c.id = $1 AND c.workspace_id = w.id;
