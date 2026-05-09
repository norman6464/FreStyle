package repository

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// MasterExerciseExampleRepository は MasterExercise に紐付く入出力例の永続化を担う。
//
// 一覧取得は `(exercise_id, order_index)` 複合インデックスで安定ソート。
// 詳細画面では特定 exercise の全 example を表示し、採点 usecase では同じ全件を
// テストケースとしてループ実行する。
type MasterExerciseExampleRepository interface {
	ListByExerciseID(exerciseID uint64) ([]domain.MasterExerciseExample, error)
	ListByExerciseIDs(exerciseIDs []uint64) (map[uint64][]domain.MasterExerciseExample, error)
}

type masterExerciseExampleRepository struct {
	db *gorm.DB
}

func NewMasterExerciseExampleRepository(db *gorm.DB) MasterExerciseExampleRepository {
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
