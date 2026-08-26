-- 章の完了記録（user_chapter_progress）のクエリ。1 行 = その (user, chapter) が完了済み。
-- completed_at / created_at に DB 既定値は無いので now() をクエリ側で明示する。id は
-- 実 DB の採番列（bigserial）に任せて INSERT では省略する。

-- name: InsertUserChapterProgressIfAbsent :execrows
-- (user, chapter) を完了として記録する。既に完了済みなら ON CONFLICT DO NOTHING で
-- 何もしない（冪等）。呼び出し側は RowsAffected>0 で初回完了かを判定する。
INSERT INTO user_chapter_progress
  (user_id, chapter_id, course_id, completed_at, created_at)
VALUES
  ($1, $2, $3, now(), now())
ON CONFLICT (user_id, chapter_id) DO NOTHING;

-- name: DeleteUserChapterProgress :exec
-- 完了記録を取り消す（行削除）。未記録でも 0 行削除でエラーにはならない。
DELETE FROM user_chapter_progress
WHERE user_id = $1 AND chapter_id = $2;

-- name: ListUserChapterProgressByUser :many
-- user の完了記録をすべて返す。ORDER BY 無しだと返却順が実行計画任せになるため PK で固定する。
SELECT id, user_id, chapter_id, course_id, completed_at, created_at
FROM user_chapter_progress
WHERE user_id = $1
ORDER BY id ASC;

-- name: CountCompletedChaptersByCourseForUser :many
-- 「現存する published 教材」の完了行のみを course_id ごとに集計する。教材削除で JOIN から
-- 落ち、非公開化は is_published で除外されるため、分子が分母(published 章数)を上回らない。
SELECT tm.course_id, COUNT(*) AS cnt
FROM user_chapter_progress ulp
JOIN course_chapters tm ON tm.id = ulp.chapter_id
WHERE ulp.user_id = $1 AND tm.is_published = TRUE
GROUP BY tm.course_id;
