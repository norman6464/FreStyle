-- name: ComputeUserStats :one
-- score_cards からユーザーの提出数(COUNT)と平均スコア(AVG)を 1 行に集計する。
-- overall_score は nullable numeric で、AVG は NULL 行を無視する。1 件も一致しなくても
-- COUNT/AVG の集計は必ず 1 行返るため :one でよい。
-- domain.UserStats.AverageScore が float64 なので AVG は double precision へ寄せる
-- （GORM 版も numeric を float64 へ Scan していた。scan 先の型を保つ）。
-- AVG が NULL（一致行なし or 全て NULL スコア）のときは COALESCE で 0 に倒す。
SELECT
  COUNT(*)::bigint AS total_sessions,
  COALESCE(AVG(overall_score), 0)::double precision AS average_score
FROM score_cards
WHERE user_id = sqlc.arg(user_id)::bigint;
