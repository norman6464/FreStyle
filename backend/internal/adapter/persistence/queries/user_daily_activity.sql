-- 日次学習サマリー（user_daily_activities）のクエリ。PK = (user_id, activity_date)。
-- created_at / updated_at 列は持たない集計テーブルなので時刻の明示は不要。

-- name: IncrementUserDailyActivity :exec
-- (user_id, activity_date) を upsert し各カウンタへ delta を原子的に加算する。
-- 行が無ければ delta の値で INSERT、あれば ON CONFLICT DO UPDATE で加算する
-- （PostgreSQL が行ロックを取るのでアプリ側のロックは不要）。
INSERT INTO user_daily_activities
  (user_id, activity_date, exercise_count, correct_count, chapter_count, note_count)
VALUES
  ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, activity_date) DO UPDATE SET
  exercise_count = user_daily_activities.exercise_count + EXCLUDED.exercise_count,
  correct_count  = user_daily_activities.correct_count  + EXCLUDED.correct_count,
  chapter_count  = user_daily_activities.chapter_count  + EXCLUDED.chapter_count,
  note_count     = user_daily_activities.note_count     + EXCLUDED.note_count;
