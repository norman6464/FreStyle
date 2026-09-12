-- name: AddPageFavorite :execrows
-- 付ける。冪等（既に付いていれば何もしない）。:execrows で影響行数を返し、呼び出し元が
-- 「今回新しく付いたか」（201 vs 204 の判定）を区別できるようにする。
INSERT INTO page_favorites (user_id, workspace_id, page_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, page_id) DO NOTHING;

-- name: RemovePageFavorite :exec
-- 外す。行の有無に関わらず成功として扱う（冪等）。
DELETE FROM page_favorites WHERE user_id = $1 AND page_id = $2;

-- name: IsPageFavorite :one
-- resolve 応答の isFavorite（★ の初期状態）に使う。
SELECT EXISTS(SELECT 1 FROM page_favorites WHERE user_id = $1 AND page_id = $2);

-- name: ListPageFavorites :many
-- そのワークスペース内の、自分のお気に入りを付けた順の新しい順に返す（ワークスペース単位。
-- GET /kb/workspaces/:workspaceSlug/favorites の裏）。可視判定はここでは行わない（呼び出し元の
-- ListPageFavoritesUseCase が CheckPagePermissionUseCase を通してからふるう。
-- ListRecentPageViewCandidates と同じ考え方 — SQL 側で絞ると、ふるいで削られる分だけ
-- 本来見えるはずの行を取りこぼす）。アーカイブ済みのページは候補にしない。
SELECT
  f.page_id,
  f.created_at,
  p.title,
  p.icon,
  p.space_id,
  s.name AS space_name
FROM page_favorites f
JOIN pages p ON p.workspace_id = f.workspace_id AND p.id = f.page_id
JOIN spaces s ON s.workspace_id = p.workspace_id AND s.id = p.space_id
WHERE f.user_id = $1 AND f.workspace_id = $2 AND p.archived_at IS NULL
ORDER BY f.created_at DESC
LIMIT 200;
