-- ページへのラベル付け外し（page_labels）のクエリ（段 13）。labels 表そのもの・
-- CRUD（Create/Find/List/Update/Delete）は ticket_label.sql をそのまま流用する
-- （ページ専用の labels は作らない。語彙をチケットと共有する）。

-- name: AddPageLabel :execrows
-- 付け外しは冪等（AddTicketLabel と同じ形 — 複合主キーの重複を無視する）。
INSERT INTO page_labels (workspace_id, page_id, label_id, created_at)
VALUES ($1, $2, $3, now())
ON CONFLICT DO NOTHING;

-- name: RemovePageLabel :execrows
DELETE FROM page_labels WHERE workspace_id = $1 AND page_id = $2 AND label_id = $3;

-- name: ListLabelsByPage :many
SELECT l.* FROM labels l
JOIN page_labels pl ON pl.workspace_id = l.workspace_id AND pl.label_id = l.id
WHERE pl.workspace_id = $1 AND pl.page_id = $2
ORDER BY l.name_key;

-- name: ListLabelsByPageIDs :many
-- page_id ごとのラベル一覧を 1 回でまとめて引く（ListLabelsByTicketIDs と同じ理由・同じ形）。
SELECT pl.page_id, l.* FROM page_labels pl
JOIN labels l ON l.workspace_id = pl.workspace_id AND l.id = pl.label_id
WHERE pl.workspace_id = sqlc.arg(workspace_id)
  AND pl.page_id IN (
    SELECT value::uuid FROM json_array_elements_text(sqlc.arg(page_ids)::json) AS t(value)
  )
ORDER BY pl.page_id, l.name_key;
