-- name: CountMembersByWorkspace :many
-- ワークスペースごとの在籍メンバー数（総数 / 有効 / trainee）を 1 クエリで集計する（運営の横断ビュー用）。
-- 論理削除済み（deleted_at IS NOT NULL）とワークスペース未所属（workspace_id IS NULL）は除外する。
-- trainee 判定は正規化後の正である role_id で行う（パラメータで渡す）。
SELECT
  workspace_id,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE is_active) AS active,
  COUNT(*) FILTER (WHERE role_id = sqlc.arg(trainee_role_id)) AS trainees
FROM users
WHERE workspace_id IS NOT NULL AND deleted_at IS NULL
GROUP BY workspace_id;
