package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// exerciseSubmissionRepository は [repository.ExerciseSubmissionRepository] の GORM 実装。
type exerciseSubmissionRepository struct {
	db *gorm.DB
}

func NewExerciseSubmissionRepository(db *gorm.DB) repository.ExerciseSubmissionRepository {
	return &exerciseSubmissionRepository{db: db}
}

func (r *exerciseSubmissionRepository) Create(ctx context.Context, submission *domain.ExerciseSubmission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

func (r *exerciseSubmissionRepository) ListByUserAndExercise(ctx context.Context, userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error) {
	var rows []domain.ExerciseSubmission
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ?", userID, exerciseID, kind).
		Order("submitted_at desc, id desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *exerciseSubmissionRepository) HasSolved(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.ExerciseSubmission{}).
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ? AND is_correct = ?", userID, exerciseID, kind, true).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *exerciseSubmissionRepository) HasAttempted(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.ExerciseSubmission{}).
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ?", userID, exerciseID, kind).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
