package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseExampleRepository は MasterExercise に紐付く入出力例の永続化を担う。
// 取得は (exercise_id, order_index) で安定ソートする。
type MasterExerciseExampleRepository interface {
	ListByExerciseID(ctx context.Context, exerciseID uint64) ([]domain.MasterExerciseExample, error)
	ListByExerciseIDs(ctx context.Context, exerciseIDs []uint64) (map[uint64][]domain.MasterExerciseExample, error)
}
