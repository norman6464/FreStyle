-- チケットの添付ファイルのメタデータ（段 4・設計 Ⅵ）。本体は Cloud Storage。

-- name: CreateTicketAttachment :one
INSERT INTO ticket_attachments
  (id, workspace_id, ticket_id, key, filename, content_type, size_bytes, uploaded_by_user_id, created_at)
VALUES (
  sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(ticket_id), sqlc.arg(key), sqlc.arg(filename),
  sqlc.arg(content_type), sqlc.arg(size_bytes), sqlc.arg(uploaded_by_user_id), now()
)
RETURNING *;

-- name: FindTicketAttachment :one
SELECT * FROM ticket_attachments WHERE workspace_id = $1 AND ticket_id = $2 AND id = $3;

-- name: ListTicketAttachments :many
SELECT * FROM ticket_attachments WHERE workspace_id = $1 AND ticket_id = $2 ORDER BY created_at;

-- name: DeleteTicketAttachment :execrows
DELETE FROM ticket_attachments WHERE workspace_id = $1 AND ticket_id = $2 AND id = $3;
