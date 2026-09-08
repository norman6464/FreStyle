-- page_templates（ページの雛形）のクエリ。

-- name: InsertPageTemplate :one
-- 雛形を 1 件作成する。id は Go 側（kbNewID）が UUIDv7 で採番して渡す。
-- name の重複（uq_page_templates_workspace_name）は Go 側で isUniqueViolation により
-- repository.ErrDuplicateTemplateName へ翻訳する（検査してから INSERT するまでの間に
-- 別の要求が同じ名前を取り得るため、一意制約を唯一の判定にする）。
INSERT INTO page_templates (id, workspace_id, space_id, name, icon, doc, created_by_user_id)
VALUES (
  sqlc.arg(id), sqlc.arg(workspace_id), sqlc.narg(space_id), sqlc.arg(name),
  sqlc.narg(icon), sqlc.arg(doc), sqlc.arg(created_by_user_id)
)
RETURNING *;

-- name: ListPageTemplates :many
-- ワークスペースの雛形一覧を name 昇順で返す。doc は一覧では使わない
-- （API 応答にも含めない）ので列に含めない。
--
-- space_id の絞り込みは UNION ではなく 1 本の WHERE で書く。sqlc.narg(space_id) に
-- NULL（Go の uuid.NullUUID{Valid: false}）を渡すと `space_id = NULL` は SQL の 3 値論理で
-- 常に NULL（偽と同じ扱い）になるため、OR の左側 `space_id IS NULL` だけが効き、
-- 結果は「space_id IS NULL の行だけ」になる。非 NULL を渡せば「space_id IS NULL の行」と
-- 「space_id = 引数の行」の両方が返る。UNION で 2 本のクエリを合成するのと同じ結果集合を、
-- 表の別名を増やさず 1 本の SELECT で得られるため、こちらを採用した。
SELECT id, workspace_id, space_id, name, icon, created_by_user_id, created_at, updated_at
FROM page_templates
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (space_id IS NULL OR space_id = sqlc.narg(space_id))
ORDER BY name;

-- name: GetPageTemplate :one
-- 雛形 1 件（doc 込み）。workspace_id まで絞ることで、他テナントの id を渡されても
-- 見つからない（＝存在しないのと同じ）ようにする。
SELECT * FROM page_templates
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id);

-- name: DeletePageTemplate :execrows
-- 雛形を削除する。影響行数で存在確認を兼ねる（0 行なら呼び出し側が
-- domain.ErrPageTemplateNotFound に翻訳する）。
DELETE FROM page_templates
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id);
