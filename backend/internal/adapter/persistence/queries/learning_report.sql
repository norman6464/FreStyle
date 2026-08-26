-- name: ListLearningReportsByUserID :many
-- 自分のレポートを期間末(period_to)降順で返す。period_to は同一期間で同着し得るため
-- id をタイブレークに置いて並びを固定する。
SELECT id, user_id, period_from, period_to, status, s3_key, created_at
FROM learning_reports
WHERE user_id = sqlc.arg(user_id)
ORDER BY period_to DESC, id DESC;

-- name: InsertLearningReport :one
-- レポートを 1 件作成する。id は採番列なので省き RETURNING で id / created_at を書き戻す。
-- created_at は DB 既定値が無いため呼び出し側が値を渡す（GORM autoCreateTime 相当。
-- ゼロなら呼び出し側で now() を入れる）。updated_at 列は持たない。
INSERT INTO learning_reports
  (user_id, period_from, period_to, status, s3_key, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at;
