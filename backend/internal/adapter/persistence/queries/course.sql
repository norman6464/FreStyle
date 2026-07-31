-- name: ListCoursesByCompany :many
-- 自社のコースを sort_order 昇順（同値時 id 昇順）で返す。
-- include_unpublished=false なら公開済み（is_published=true）のみに絞る。
SELECT * FROM courses
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_unpublished)::bool OR is_published = TRUE)
ORDER BY sort_order ASC, id ASC;

-- name: GetCourseByID :one
-- 内部 ID で 1 件取得（存在しなければ sql.ErrNoRows）。
SELECT * FROM courses
WHERE id = $1;