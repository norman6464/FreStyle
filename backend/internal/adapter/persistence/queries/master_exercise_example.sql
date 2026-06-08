-- name: ListMasterExerciseExamplesByExerciseID :many
-- 1 問の入力例 / 期待出力例を表示・採点順（order_index, id）で返す。
SELECT * FROM master_exercise_examples
WHERE exercise_id = $1
ORDER BY order_index ASC, id ASC;

-- 複数 exercise_id をまとめて取る IN 句のスライス展開は、postgres × database/sql モードでは
-- sqlc.slice が正しく生成されない（pgx モードでは = ANY($1) で素直に書ける）。
-- バッチ取得は GORM 撤去フェーズで接続を pgx へ寄せる際にここへ追加する。
