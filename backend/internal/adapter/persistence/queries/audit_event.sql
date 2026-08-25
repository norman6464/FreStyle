-- name: ListRecentAuditEvents :many
SELECT * FROM audit_events
ORDER BY created_at DESC, id DESC
LIMIT $1;