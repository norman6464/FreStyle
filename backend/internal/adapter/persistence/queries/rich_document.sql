-- name: GetRichDocumentByID :one
-- ID で 1 件引く（論理削除は除外）。無ければ sql.ErrNoRows。
-- 誰が読めるか（テナント境界）は domain.RichDocument.CanBeReadBy が決めるので、ここでは
-- 可視性条件を足さない（SQL 側へ写経すると片方だけ直したときに食い違う）。
SELECT id, owner_id, company_id, kind, title, is_public, schema_version, doc, revision, created_at, updated_at, deleted_at
FROM rich_documents
WHERE id = sqlc.arg(id) AND deleted_at IS NULL;

-- name: InsertRichDocument :one
-- 文書を 1 件作成する。id はアプリ採番の uuid なので明示的に渡す。
-- created_at / updated_at は DB 既定値が無いため呼び出し側が値を渡す（autoTime 相当。
-- ゼロなら呼び出し側で now() を入れる）。schema_version / revision は 0 のとき既定 1 を当てる
-- （GORM の `default:1` タグと同じ挙動。RETURNING で確定値を書き戻す）。
INSERT INTO rich_documents
  (id, owner_id, company_id, kind, title, is_public, schema_version, doc, revision, created_at, updated_at)
VALUES (
  sqlc.arg(id),
  sqlc.arg(owner_id),
  sqlc.narg(company_id),
  sqlc.arg(kind),
  sqlc.arg(title),
  sqlc.arg(is_public),
  COALESCE(NULLIF(sqlc.arg(schema_version)::int, 0), 1),
  sqlc.arg(doc),
  COALESCE(NULLIF(sqlc.arg(revision)::int, 0), 1),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
)
RETURNING schema_version, revision, created_at, updated_at;

-- name: UpdateRichDocumentWithRevision :one
-- 楽観ロック付き更新。expected_revision が現在値と一致し論理削除されていない行だけを更新し、
-- revision を +1・updated_at を now() へ進める。0 行なら sql.ErrNoRows（呼び出し側が存在確認で
-- 404 / 409 を切り分ける）。RETURNING で更新後の行全体を返し struct に反映する。
UPDATE rich_documents SET
  title          = sqlc.arg(title),
  is_public      = sqlc.arg(is_public),
  schema_version = sqlc.arg(schema_version),
  doc            = sqlc.arg(doc),
  revision       = revision + 1,
  updated_at     = now()
WHERE id = sqlc.arg(id) AND revision = sqlc.arg(expected_revision) AND deleted_at IS NULL
RETURNING id, owner_id, company_id, kind, title, is_public, schema_version, doc, revision, created_at, updated_at, deleted_at;

-- name: SoftDeleteRichDocument :execrows
-- owner を条件に論理削除する（他人の文書は消せない）。RowsAffected=0 は対象なし（404）。
UPDATE rich_documents SET
  deleted_at = now()
WHERE id = sqlc.arg(id) AND owner_id = sqlc.arg(owner_id) AND deleted_at IS NULL;

-- name: ListRichDocumentsByOwner :many
-- owner の文書を更新日降順（同時刻は id 降順）で返す（論理削除は除外）。
-- kind が空文字なら絞り込まない。一覧用途のため doc(jsonb) 本体は読み込まない（軽量サマリ）。
SELECT id, owner_id, company_id, kind, title, is_public, schema_version, revision, created_at, updated_at
FROM rich_documents
WHERE owner_id = sqlc.arg(owner_id)
  AND deleted_at IS NULL
  AND (sqlc.arg(kind)::text = '' OR kind = sqlc.arg(kind))
ORDER BY updated_at DESC, id DESC;
