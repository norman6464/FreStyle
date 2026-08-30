-- name: ListCoursesByWorkspace :many
-- 自社のコースを sort_order 昇順（同値時 id 昇順）で返す。
-- include_unpublished=false なら公開済み（is_published=true）のみに絞る。
--
-- FRESTYLE-400（段4横展開）: company_id 直読みから workspace_id 経由へ切り替え済み。
SELECT * FROM courses
WHERE workspace_id = sqlc.arg(workspace_id)
  AND (sqlc.arg(include_unpublished)::bool OR is_published = TRUE)
ORDER BY sort_order ASC, id ASC;

-- name: GetCourseByID :one
-- 内部 ID で 1 件取得（存在しなければ sql.ErrNoRows）。
SELECT * FROM courses
WHERE id = $1;
-- name: InsertCourse :one
-- コースを 1 件作成する。id は採番列なので省き RETURNING で id / sort_order / created_at / updated_at を
-- 書き戻す。created_at / updated_at は DB 既定値が無いため呼び出し側が値を渡す（autoTime 相当。
-- ゼロなら呼び出し側で now() を入れる）。sort_order は 0 のとき既定 100 を当てる
-- （GORM の `default:100` タグと同じ挙動。RETURNING で確定値を書き戻す）。
--
-- workspace_id は company_id からその場で引く（FRESTYLE-399。dual-write を起動時
-- バックフィルだけに任せると、次の起動までのあいだに作ったコースが workspace_id 経由の
-- 一覧から漏れる。FRESTYLE-397 の InsertUser と同じ理由・同じ形）。
INSERT INTO courses
  (company_id, workspace_id, created_by_user_id, title, description, category, language, sort_order, is_published, created_at, updated_at)
VALUES (
  sqlc.arg(company_id),
  (SELECT c.workspace_id FROM companies c WHERE c.id = sqlc.arg(company_id)),
  sqlc.arg(created_by_user_id),
  sqlc.arg(title),
  sqlc.arg(description),
  sqlc.arg(category),
  sqlc.arg(language),
  COALESCE(NULLIF(sqlc.arg(sort_order)::bigint, 0), 100),
  sqlc.arg(is_published),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
)
RETURNING id, sort_order, created_at, updated_at;

-- name: UpdateCourse :one
-- コースを部分更新する。書くのは title / description / sort_order / is_published の 4 列だけで、
-- created_by_user_id / company_id / category / language / created_at は不変（GORM の Updates(map) と同じ）。
-- updated_at は now() へ進めて RETURNING で書き戻す（autoUpdateTime 相当）。
UPDATE courses SET
  title        = sqlc.arg(title),
  description  = sqlc.arg(description),
  sort_order   = sqlc.arg(sort_order),
  is_published = sqlc.arg(is_published),
  updated_at   = now()
WHERE id = sqlc.arg(id)
RETURNING updated_at;

-- name: DeleteCourse :execrows
-- コースを物理削除する（courses は soft delete 列を持たない）。
--
-- :exec ではなく :execrows にしている理由:
--   :exec は「SQL がエラーなく流れたか」しか返さない。DELETE は 1 行も一致しなくても
--   成功なので、存在しない id を渡しても呼び出し側には成功として見える。
--   :execrows は実際に消えた行数（RowsAffected）を返すので、repository が 0 行を
--   「対象なし」として domain.ErrNotFound に翻訳できる。
DELETE FROM courses
WHERE id = sqlc.arg(id);
