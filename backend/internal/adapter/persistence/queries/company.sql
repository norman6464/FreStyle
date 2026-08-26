-- name: ListCompanies :many
-- 企業一覧（名前昇順）。name に一意制約は無いので同名企業の順序を固定する id ASC を付ける。
SELECT * FROM companies
ORDER BY name ASC, id ASC;

-- name: GetCompanyByID :one
-- ID で企業を 1 件取得。
SELECT * FROM companies
WHERE id = $1;

-- name: UpdateCompanyAiChatEnabled :execrows
-- trainee への AI チャット許可を更新する。0 件なら対象の会社が存在しない
-- （呼び出し側が not-found にする）。判定は UpdateCompanyActive と同じく companies 側の
-- 件数で行う（写し先の workspaces の件数を見ると、まだ紐付いていない会社が not-found に化ける）。
--
-- :exec ではなく :execrows にしている理由:
--   :exec は「SQL がエラーなく流れたか」しか返さない。UPDATE は 1 行も一致しなくても
--   成功なので、存在しない会社の設定を書こうとしても呼び出し側には成功として見え、
--   画面には切り替えたはずの設定が反映されたように表示される。
UPDATE companies SET ai_chat_enabled_for_trainees = $2, updated_at = now() WHERE id = $1;

-- name: UpdateCompanyActive :execrows
-- 会社アカウントの有効/無効を更新する。0 件なら対象が存在しない（呼び出し側が not-found にする）。
-- 判定は companies 側の件数で行う。写し先（workspaces）の件数を見ると、まだ紐付いていない
-- 会社が not-found に化ける。
UPDATE companies SET is_active = $2, updated_at = now() WHERE id = $1;

-- name: MirrorCompanySettingsToWorkspace :exec
-- 会社のテナント設定を、対応する workspaces 行へ写す。
-- ai_chat_enabled_for_trainees / is_active は最終的に workspaces の列になる。今は companies が
-- 正本で、workspaces 側はその写し。設定を書く経路が増えても写し忘れないよう、2 列まとめて
-- companies から写すこの 1 か所に集約する。まだワークスペースに紐付いていない会社
-- （workspace_id IS NULL）は 0 件更新で、写す先が無い。
UPDATE workspaces w
SET ai_chat_enabled_for_trainees = c.ai_chat_enabled_for_trainees,
    is_active = c.is_active,
    updated_at = now()
FROM companies c
WHERE c.id = $1 AND c.workspace_id = w.id;
