-- name: InsertExerciseSubmission :one
-- 提出履歴を 1 件追加する（append-only）。id は採番列なので省き RETURNING で書き戻す。
-- submitted_at は DB 既定値が無く autoTime 対象でもないため呼び出し側（usecase）が必ず値を渡す。
-- stdout / stderr は nullable 列だが、GORM 版は非ポインタ string を常に書いていたため
-- 空文字も NULL ではなく '' として保存する（Valid=true で渡す）。
INSERT INTO exercise_submissions
  (user_id, exercise_kind, exercise_id, submitted_code, stdout, stderr, exit_code, is_correct, submitted_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id;

-- name: ListSubmissionsByUserAndExercise :many
-- user × (kind, exercise_id) の提出履歴を新しい順に返す。
-- submitted_at は同着し得るため id 降順をタイブレークに置いて並びを固定する。
SELECT id, user_id, exercise_kind, exercise_id, submitted_code, stdout, stderr, exit_code, is_correct, submitted_at
FROM exercise_submissions
WHERE user_id = sqlc.arg(user_id)
  AND exercise_id = sqlc.arg(exercise_id)
  AND exercise_kind = sqlc.arg(exercise_kind)
ORDER BY submitted_at DESC, id DESC;

-- name: ExistsCorrectSubmission :one
-- user が (kind, exercise_id) を 1 回でも is_correct=true で解いたかを返す。
SELECT EXISTS(
  SELECT 1 FROM exercise_submissions
  WHERE user_id = sqlc.arg(user_id)
    AND exercise_id = sqlc.arg(exercise_id)
    AND exercise_kind = sqlc.arg(exercise_kind)
    AND is_correct = TRUE
) AS solved;

-- name: ExistsSubmission :one
-- user が (kind, exercise_id) に 1 回でも提出したかを返す。
SELECT EXISTS(
  SELECT 1 FROM exercise_submissions
  WHERE user_id = sqlc.arg(user_id)
    AND exercise_id = sqlc.arg(exercise_id)
    AND exercise_kind = sqlc.arg(exercise_kind)
) AS attempted;
