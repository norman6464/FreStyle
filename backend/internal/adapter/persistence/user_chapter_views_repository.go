package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

type userChapterViewRepository struct {
	db *gorm.DB
}

// NewUserChapterViewRepository は UserChapterViewRepository の GORM 実装を返す。
func NewUserChapterViewRepository(db *gorm.DB) repository.UserChapterViewRepository {
	return &userChapterViewRepository{db: db}
}

// UpsertView は章閲覧を記録する。初回 INSERT、2 回目以降は last_viewed_at と view_count を更新する。
func (r *userChapterViewRepository) UpsertView(
	ctx context.Context,
	userID, teachingMaterialID, courseID uint64,
) error {
	sql := `
INSERT INTO user_chapter_views
  (user_id, teaching_material_id, course_id, first_viewed_at, last_viewed_at, view_count)
VALUES
  (?, ?, ?, NOW(), NOW(), 1)
ON CONFLICT (user_id, teaching_material_id) DO UPDATE SET
  last_viewed_at = NOW(),
  view_count     = user_chapter_views.view_count + 1
`
	return r.db.WithContext(ctx).Exec(sql, userID, teachingMaterialID, courseID).Error
}

func (r *userChapterViewRepository) ListRecentByUser(
	ctx context.Context,
	userID uint64,
	limit int,
) ([]domain.UserChapterView, error) {
	var rows []domain.UserChapterView
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("last_viewed_at DESC").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}
