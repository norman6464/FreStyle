-- name: ListChaptersByCompany :many
-- 会社内の全教材（章）を更新日降順で返す backward-compat 用。
-- include_unpublished=false なら公開済み（is_published=true）のみに絞る。
SELECT id, company_id, course_id, created_by_user_id, title, doc, revision, schema_version, sort_order, is_published, created_at, updated_at
FROM course_chapters
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_unpublished)::bool OR is_published = TRUE)
ORDER BY updated_at DESC, id DESC;

-- name: ListChaptersByCourse :many
-- コース内の章を sort_order 昇順（同値時 id 昇順）で返す。
-- 一覧は本文（doc・jsonb）を返さない（章ごとに重く、全章を先読みすると非効率）。
SELECT id, company_id, course_id, created_by_user_id, title, sort_order, is_published, created_at, updated_at
FROM course_chapters
WHERE course_id = sqlc.arg(course_id)
  AND (sqlc.arg(include_unpublished)::bool OR is_published = TRUE)
ORDER BY sort_order ASC, id ASC;

-- name: GetChapterByID :one
-- 単一教材を返す（本文 doc を含む）。存在しなければ sql.ErrNoRows。
SELECT id, company_id, course_id, created_by_user_id, title, doc, revision, schema_version, sort_order, is_published, created_at, updated_at
FROM course_chapters
WHERE id = sqlc.arg(id);

-- name: CountChaptersByCourseForCompany :many
-- course_id ごとの教材件数を 1 クエリで集計する。
-- include_unpublished=false（trainee 相当）は published のみ数える。
SELECT course_id, COUNT(*) AS cnt
FROM course_chapters
WHERE company_id = sqlc.arg(company_id)
  AND (sqlc.arg(include_unpublished)::bool OR is_published = TRUE)
GROUP BY course_id;

-- name: InsertChapter :one
-- 教材（章）を 1 件作成する。id は採番列なので省き RETURNING で書き戻す。doc は作成時 NULL。
-- created_at / updated_at は DB 既定値が無いため呼び出し側が値を渡す（autoTime 相当）。
-- revision / schema_version は 0 のとき既定 1、sort_order は 0 のとき既定 100 を当てる
-- （GORM の `default:` タグと同じ挙動。RETURNING で確定値を書き戻す）。
INSERT INTO course_chapters
  (company_id, course_id, created_by_user_id, title, revision, schema_version, sort_order, is_published, created_at, updated_at)
VALUES (
  sqlc.arg(company_id),
  sqlc.arg(course_id),
  sqlc.arg(created_by_user_id),
  sqlc.arg(title),
  COALESCE(NULLIF(sqlc.arg(revision)::bigint, 0), 1),
  COALESCE(NULLIF(sqlc.arg(schema_version)::bigint, 0), 1),
  COALESCE(NULLIF(sqlc.arg(sort_order)::bigint, 0), 100),
  sqlc.arg(is_published),
  sqlc.arg(created_at),
  sqlc.arg(updated_at)
)
RETURNING id, revision, schema_version, sort_order, created_at, updated_at;

-- name: UpdateChapter :one
-- 教材のメタ情報を部分更新する。書くのは title / sort_order / is_published の 3 列だけで、
-- created_by_user_id / company_id / course_id / doc / revision は不変（GORM の Updates(map) と同じ）。
-- updated_at は now() へ進めて RETURNING で書き戻す（autoUpdateTime 相当）。
UPDATE course_chapters SET
  title        = sqlc.arg(title),
  sort_order   = sqlc.arg(sort_order),
  is_published = sqlc.arg(is_published),
  updated_at   = now()
WHERE id = sqlc.arg(id)
RETURNING updated_at;

-- name: UpdateChapterDocWithRevision :one
-- リッチ本文（tiptap JSON）を楽観ロックで更新する。expected_revision が現在値と一致した行だけを
-- 更新し、revision を +1・updated_at を now() へ進める。0 行なら sql.ErrNoRows（呼び出し側が
-- 存在確認で 404 / 409 を切り分ける）。RETURNING で更新後の行全体を返す。
UPDATE course_chapters SET
  doc        = sqlc.arg(doc),
  revision   = revision + 1,
  updated_at = now()
WHERE id = sqlc.arg(id) AND revision = sqlc.arg(expected_revision)
RETURNING id, company_id, course_id, created_by_user_id, title, doc, revision, schema_version, sort_order, is_published, created_at, updated_at;

-- name: DeleteChapter :exec
-- 教材を物理削除する（course_chapters は soft delete 列を持たない）。
DELETE FROM course_chapters WHERE id = sqlc.arg(id);

-- name: DeleteChaptersByCourse :exec
-- コース削除時の cascade 用に配下教材を全削除する（FK に頼らず明示削除）。
DELETE FROM course_chapters WHERE course_id = sqlc.arg(course_id);
