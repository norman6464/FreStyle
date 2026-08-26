-- name: ListMasterExerciseExamplesByExerciseID :many
-- 1 問の入力例 / 期待出力例を表示・採点順（order_index, id）で返す。
SELECT * FROM master_exercise_examples
WHERE exercise_id = $1
ORDER BY order_index ASC, id ASC;

-- name: ListMasterExerciseExamplesByExerciseIDs :many
-- 複数 exercise_id をまとめて取り、呼び出し側で exercise_id ごとに map 化する（N+1 回避）。
-- 表示・採点順を安定させるため (exercise_id, order_index, id) 昇順で返す。
--
-- IN 句のスライス展開に sqlc.slice / = ANY($1::bigint[]) を使うと、database/sql モードでは
-- 配列を pq.Array で包む生成になり lib/pq への依存が増える（このリポジトリの driver は pgx）。
-- 依存を増やさないため、id 群は 1 個の json 配列パラメータで渡し、json_array_elements_text で
-- 展開してから bigint へ落とす。
SELECT id, exercise_id, order_index, input_text, expected_output, created_at, updated_at
FROM master_exercise_examples
WHERE exercise_id IN (
  SELECT value::bigint FROM json_array_elements_text(sqlc.arg(exercise_ids)::json) AS t(value)
)
ORDER BY exercise_id ASC, order_index ASC, id ASC;
