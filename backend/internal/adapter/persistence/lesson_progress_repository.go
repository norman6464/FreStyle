package persistence

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// lessonProgressRepository は [repository.LessonProgressRepository] の GORM 実装。
type lessonProgressRepository struct {
	db *gorm.DB
}

func NewLessonProgressRepository(db *gorm.DB) repository.LessonProgressRepository {
	return &lessonProgressRepository{db: db}
}

func (r *lessonProgressRepository) MarkCompleted(ctx context.Context, userID, materialID, courseID uint64) (bool, error) {
	row := &domain.UserLessonProgress{
		UserID:             userID,
		TeachingMaterialID: materialID,
		CourseID:           courseID,
		CompletedAt:        time.Now(),
	}
	// (user_id, teaching_material_id) が衝突したら何もしない（冪等）。RowsAffected で初回かを判定。
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "teaching_material_id"}},
		DoNothing: true,
	}).Create(row)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *lessonProgressRepository) MarkIncomplete(ctx context.Context, userID, materialID uint64) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND teaching_material_id = ?", userID, materialID).
		Delete(&domain.UserLessonProgress{}).Error
}

func (r *lessonProgressRepository) ListByUser(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error) {
	var rows []domain.UserLessonProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
