package persistence

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// masterExerciseExampleRepository は [repository.MasterExerciseExampleRepository] の GORM 実装。
type masterExerciseExampleRepository struct {
	db *gorm.DB
}

// NewMasterExerciseExampleRepository は GORM ベース の
// [repository.MasterExerciseExampleRepository] を 返す。
func NewMasterExerciseExampleRepository(db *gorm.DB) repository.MasterExerciseExampleRepository {
	return &masterExerciseExampleRepository{db: db}
}

func (r *masterExerciseExampleRepository) ListByExerciseID(exerciseID uint64) ([]domain.MasterExerciseExample, error) {
	var examples []domain.MasterExerciseExample
	if err := r.db.
		Where("exercise_id = ?", exerciseID).
		Order("order_index asc, id asc").
		Find(&examples).Error; err != nil {
		return nil, err
	}
	return examples, nil
}

// ListByExerciseIDs は複数 exercise_id をまとめて取得し、 exercise_id ごとに
// グルーピングした map を返す。 リストページで N+1 を避けるため。
func (r *masterExerciseExampleRepository) ListByExerciseIDs(exerciseIDs []uint64) (map[uint64][]domain.MasterExerciseExample, error) {
	result := make(map[uint64][]domain.MasterExerciseExample, len(exerciseIDs))
	if len(exerciseIDs) == 0 {
		return result, nil
	}
	var examples []domain.MasterExerciseExample
	if err := r.db.
		Where("exercise_id IN ?", exerciseIDs).
		Order("exercise_id asc, order_index asc, id asc").
		Find(&examples).Error; err != nil {
		return nil, err
	}
	for _, ex := range examples {
		result[ex.ExerciseID] = append(result[ex.ExerciseID], ex)
	}
	return result, nil
}
