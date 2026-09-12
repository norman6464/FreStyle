package domain

import "time"

// PageFavorite は「お気に入り一覧」1 件の投影（page_favorites と pages/spaces の結合結果）。
// page_views と同じく、専用の永続化 domain 型は持たず、この投影自体が唯一の使い道。
type PageFavorite struct {
	PageID    string
	Title     string
	Icon      *PageIcon
	SpaceID   string
	SpaceName string
	CreatedAt time.Time
}
