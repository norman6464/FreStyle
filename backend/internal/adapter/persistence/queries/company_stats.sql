-- name: CountMembersByCompany :many
-- 会社ごとの在籍メンバー数（総数 / 有効 / trainee）を 1 クエリで集計する（運営の横断ビュー用）。
-- 論理削除済み（deleted_at IS NOT NULL）と会社未所属（company_id IS NULL）は除外する。
-- trainee 判定は正規化後の正である role_id で行う（パラメータで渡す）。
-- WHERE で company_id IS NOT NULL に絞っているので COALESCE は NULL を返さない
-- （型を非 NULL の bigint に確定させ、詰め替えを綺麗にするための cast）。
SELECT
  COALESCE(company_id, 0)::bigint AS company_id,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE is_active) AS active,
  COUNT(*) FILTER (WHERE role_id = sqlc.arg(trainee_role_id)) AS trainees
FROM users
WHERE company_id IS NOT NULL AND deleted_at IS NULL
GROUP BY company_id;

-- name: CountMembersByWorkspace :many
-- ワークスペースごとの在籍メンバー数（総数 / 有効 / trainee）を 1 クエリで集計する。
-- CountMembersByCompany のワークスペース版（company_id を DROP する段5準備）。論理削除済み
-- （deleted_at IS NOT NULL）とワークスペース未所属（workspace_id IS NULL）は除外する。
-- trainee 判定は正規化後の正である role_id で行う（パラメータで渡す）。
-- workspace_id は非 NULL に絞っているため COALESCE は不要（company_id 版と違い uuid に
-- ダミー値の代替が無いため、company_id=0 相当の詰め替えはしない）。
SELECT
  workspace_id,
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE is_active) AS active,
  COUNT(*) FILTER (WHERE role_id = sqlc.arg(trainee_role_id)) AS trainees
FROM users
WHERE workspace_id IS NOT NULL AND deleted_at IS NULL
GROUP BY workspace_id;
