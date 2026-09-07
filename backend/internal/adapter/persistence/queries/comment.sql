-- ページ全体へのコメント（comment_threads / comments）のクエリ。FRESTYLE-432 段 2。
--
-- 段 2 の時点では書き込み経路が「本文」だけを受け取るので、block_id / anchor_from /
-- anchor_to / quote への INSERT は無い（常に NULL のまま作られる。段 3 で足す）。

-- name: CreateCommentThread :one
INSERT INTO comment_threads (id, workspace_id, page_id, created_by_user_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CreateComment :one
INSERT INTO comments (id, thread_id, author_user_id, body)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetCommentThread :one
-- workspace_id / page_id まで絞ることで、他ページ・他テナントの thread_id を渡されても
-- 見つからない（＝存在しないのと同じ）ようにする。
SELECT * FROM comment_threads
WHERE workspace_id = $1 AND page_id = $2 AND id = $3;

-- name: ListCommentThreadsByPage :many
SELECT * FROM comment_threads
WHERE workspace_id = $1 AND page_id = $2
ORDER BY created_at ASC, id ASC;

-- name: ListCommentsByThreadIDs :many
-- id 群は json 配列 1 個のパラメータで渡し、json_array_elements_text で展開する
-- （= ANY(sqlc.arg(ids)::uuid[]) は database/sql モードの sqlc で pq.Array() 依存になり
-- ビルドが壊れる。理由と前例は internal/adapter/persistence/queries/knowledge_base.sql の
-- ListExistingBlockIDsAmong のコメントを参照）。
SELECT * FROM comments
WHERE thread_id IN (
  SELECT value::uuid FROM json_array_elements_text(sqlc.arg(thread_ids)::json) AS t(value)
)
ORDER BY created_at ASC, id ASC;

-- name: ResolveCommentThread :one
UPDATE comment_threads
SET resolved_at = now(), resolved_by_user_id = sqlc.arg(resolved_by_user_id)::bigint, updated_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND page_id = sqlc.arg(page_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: ReopenCommentThread :one
UPDATE comment_threads
SET resolved_at = NULL, resolved_by_user_id = NULL, updated_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND page_id = sqlc.arg(page_id) AND id = sqlc.arg(id)
RETURNING *;
