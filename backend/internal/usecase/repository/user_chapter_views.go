package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// UserChapterViewRepository は章閲覧記録の永続化 port。
type UserChapterViewRepository interface {
	// UpsertView は (user_id, chapter_id) の行を upsert する。
	// 初回: INSERT。2回目以降: last_viewed_at を現在時刻に更新し view_count を +1。
	UpsertView(ctx context.Context, userID, teachingMaterialID, courseID uint64) error

	// GetLastViewedByUserAndCourse は user がコース内で最後に閲覧した 1 件を返す。
	// コース詳細のレジューム(続きから表示)用。履歴が無い場合は (nil, nil)。
	GetLastViewedByUserAndCourse(ctx context.Context, userID, courseID uint64) (*domain.UserChapterView, error)
}
