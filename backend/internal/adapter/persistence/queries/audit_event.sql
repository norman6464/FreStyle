-- name: ListRecentAuditEvents :many
-- 監査ログを新しい順で返す。created_at は一意でないため id をタイブレークに置く。
-- row_limit は bigint で受ける（LIMIT の型に合わせ、Go 側で int32 へ縮めないため）。
SELECT id, actor_id, actor_email, actor_role, action, target_id, created_at
FROM audit_events
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit)::bigint;

-- name: InsertAuditEvent :one
-- 監査ログを 1 件記録する。id は採番列なので省き RETURNING で id / created_at を書き戻す。
-- created_at は DB 既定値が無いため呼び出し側が値を渡す（GORM の autoCreateTime 相当。
-- ゼロなら呼び出し側が now() を入れる。過去分の取り込みで明示した時刻はそのまま保つ）。
INSERT INTO audit_events (actor_id, actor_email, actor_role, action, target_id, created_at)
VALUES (
  sqlc.arg(actor_id),
  sqlc.arg(actor_email),
  sqlc.arg(actor_role),
  sqlc.arg(action),
  sqlc.arg(target_id),
  sqlc.arg(created_at)
)
RETURNING id, created_at;
