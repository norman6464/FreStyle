-- name: UpsertPageView :exec
-- 人 × ページの「最後に見た日時」を 1 行だけ持つ（開くたびに upsert。来訪ごとに行を積まない）。
-- 衝突キー (user_id, page_id) は両方とも所有者列だが、workspace_id もこの表に乗っている
-- 所有者列のため、DO UPDATE の WHERE で workspace_id も絞る（queries_static_check_test.go が
-- 求める多層防御。blocks の UpsertBlock と同じ形。knowledge_base.sql 冒頭のコメント参照）。
INSERT INTO page_views (user_id, workspace_id, page_id, viewed_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (user_id, page_id) DO UPDATE SET viewed_at = now()
WHERE page_views.workspace_id = EXCLUDED.workspace_id;

-- name: CountPageViews :one
-- そのページの閲覧数。upsert 型の表なので「延べ回数」ではなく「見たことのある人数」になる
-- （page_views.sql の表コメント参照）。
SELECT count(*) FROM page_views WHERE page_id = $1;

-- name: ListRecentPageViewCandidates :many
-- 自分の「最近見たページ」候補を viewed_at の新しい順に返す。可視判定はここでは行わない
-- （呼び出し元の ListMyRecentPagesUseCase が CheckPagePermissionUseCase を通してからふるう。
-- SearchViewablePagesUseCase と同じ考え方 — SQL 側で件数を絞ると、ふるいで削られる分だけ
-- 本来見えるはずの行を取りこぼす）。アーカイブ済みのページは候補にしない。
SELECT
  v.page_id,
  v.viewed_at,
  p.workspace_id,
  p.title,
  p.icon,
  p.space_id,
  s.name AS space_name,
  w.slug AS workspace_slug
FROM page_views v
JOIN pages p ON p.workspace_id = v.workspace_id AND p.id = v.page_id
JOIN spaces s ON s.workspace_id = p.workspace_id AND s.id = p.space_id
JOIN workspaces w ON w.id = p.workspace_id
WHERE v.user_id = $1 AND p.archived_at IS NULL
ORDER BY v.viewed_at DESC
LIMIT 30;
