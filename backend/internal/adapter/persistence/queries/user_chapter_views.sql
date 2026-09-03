-- 章閲覧記録（user_chapter_views）のクエリ。PK = (user_id, chapter_id)。
-- first_viewed_at / last_viewed_at に DB 既定値は無いので、時刻はクエリ側で now() を明示する。

-- name: UpsertUserChapterView :exec
-- 章閲覧を記録する。初回は view_count=1 で first/last をセット、2 回目以降は
-- last_viewed_at を now() に、view_count を +1、course_id を最新値へ更新する。
-- first_viewed_at は DO UPDATE で触らないので初回の値が保持される。
INSERT INTO user_chapter_views
  (user_id, chapter_id, course_id, first_viewed_at, last_viewed_at, view_count)
VALUES
  ($1, $2, $3, now(), now(), 1)
ON CONFLICT (user_id, chapter_id) DO UPDATE SET
  course_id      = EXCLUDED.course_id,
  last_viewed_at = now(),
  view_count     = user_chapter_views.view_count + 1;

-- name: GetLastViewedUserChapterViewByCourse :one
-- (user, course) の閲覧記録から last_viewed_at 最大の 1 件を返す（コース詳細のレジューム用）。
-- 同時刻の章が複数あっても返る 1 件がぶれないよう chapter_id をタイブレークに置く。
SELECT user_id, chapter_id, course_id, first_viewed_at, last_viewed_at, view_count
FROM user_chapter_views
WHERE user_id = $1 AND course_id = $2
ORDER BY last_viewed_at DESC, chapter_id DESC
LIMIT 1;
