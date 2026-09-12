package domain

import "time"

// RecentPage は「最近見たページ」1 件の投影（page_views と pages/spaces/workspaces の
// 結合結果）。対応する専用の永続化 domain 型（PageView 等）は持たない — page_views は
// 「最後に見た日時」を人 × ページで 1 行だけ持つ upsert 型の表で、この投影自体が唯一の
// 使い道のため。WorkspaceID は可視判定（CheckPagePermissionUseCase）に使う内部値で、
// HTTP 応答には載せない（handler 側の response 型が絞る）。
type RecentPage struct {
	PageID        string
	WorkspaceID   string
	WorkspaceSlug string
	Title         string
	Icon          *PageIcon
	SpaceID       string
	SpaceName     string
	ViewedAt      time.Time
}
