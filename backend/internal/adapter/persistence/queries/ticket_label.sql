-- ラベル（labels）とチケットへの付け外し（ticket_labels）のクエリ（段 4・設計 Ⅵ）。

-- =============================================================================
-- labels
-- =============================================================================

-- name: CreateLabel :one
INSERT INTO labels (id, workspace_id, space_id, name, color, created_at, updated_at)
VALUES (sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(space_id), sqlc.arg(name), sqlc.arg(color), now(), now())
RETURNING *;

-- name: FindLabel :one
SELECT * FROM labels WHERE workspace_id = $1 AND id = $2;

-- name: ListLabels :many
SELECT * FROM labels WHERE workspace_id = $1 AND space_id = $2 ORDER BY name_key;

-- name: UpdateLabel :one
-- space_id で絞るのは、呼び出し側が権限を確かめた相手（URL のスペース）と
-- 実際に書き換える行を必ず一致させるため。usecase 側でも同じ突き合わせをしているが、
-- 新しい呼び出し元がその一手を忘れても、ここで 0 行に落ちて黙って通ることはない。
UPDATE labels
SET name = sqlc.arg(name), color = sqlc.arg(color), updated_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND space_id = sqlc.arg(space_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: DeleteLabel :execrows
-- ticket_labels は ON DELETE CASCADE で一緒に消える。
-- space_id で絞る理由は UpdateLabel と同じ。
DELETE FROM labels WHERE workspace_id = $1 AND space_id = $2 AND id = $3;

-- =============================================================================
-- ticket_labels
-- =============================================================================

-- name: AddTicketLabel :execrows
-- 付け外しは冪等（ticket_comment_reactions と同じ形 — 複合主キーの重複を無視する）。
INSERT INTO ticket_labels (workspace_id, ticket_id, label_id, created_at)
VALUES ($1, $2, $3, now())
ON CONFLICT DO NOTHING;

-- name: RemoveTicketLabel :execrows
DELETE FROM ticket_labels WHERE workspace_id = $1 AND ticket_id = $2 AND label_id = $3;

-- name: ListLabelsByTicket :many
SELECT l.* FROM labels l
JOIN ticket_labels tl ON tl.workspace_id = l.workspace_id AND tl.label_id = l.id
WHERE tl.workspace_id = $1 AND tl.ticket_id = $2
ORDER BY l.name_key;

-- name: ListLabelsByTicketIDs :many
-- ticket_id ごとのラベル一覧を 1 回でまとめて引く（一覧画面の N+1 を避ける）。
-- ticket_id 群は json 配列 1 個のパラメータで渡す（comment.sql の ListCommentsByThreadIDs と
-- 同じ作法 — `= ANY(...)::uuid[]` は database/sql モードの sqlc で pq.Array() 依存になり
-- ビルドが壊れるため使わない）。
SELECT tl.ticket_id, l.* FROM ticket_labels tl
JOIN labels l ON l.workspace_id = tl.workspace_id AND l.id = tl.label_id
WHERE tl.workspace_id = sqlc.arg(workspace_id)
  AND tl.ticket_id IN (
    SELECT value::uuid FROM json_array_elements_text(sqlc.arg(ticket_ids)::json) AS t(value)
  )
ORDER BY tl.ticket_id, l.name_key;
