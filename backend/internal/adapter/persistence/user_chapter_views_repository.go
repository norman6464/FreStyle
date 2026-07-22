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
  (user_id, chapter_id, course_id, first_viewed_at, last_viewed_at, view_count)
VALUES
  (?, ?, ?, NOW(), NOW(), 1)
ON CONFLICT (user_id, chapter_id) DO UPDATE SET
  course_id      = EXCLUDED.course_id,
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

// GetLastViewedByUserAndCourse は (user, course) の閲覧記録から last_viewed_at 最大の 1 件を返す。
// 履歴なしはエラーではなく (nil, nil)(「初めて開くコース」は正常系のため)。
func (r *userChapterViewRepository) GetLastViewedByUserAndCourse(
	ctx context.Context,
	userID, courseID uint64,
) (*domain.UserChapterView, error) {
	var rows []domain.UserChapterView
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND course_id = ?", userID, courseID).
		Order("last_viewed_at DESC").
		Limit(1).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return &rows[0], nil
}
