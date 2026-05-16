package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseExampleRepository は MasterExercise に紐付く入出力例の永続化を担う。
//
// 一覧取得は `(exercise_id, order_index)` 複合インデックスで安定ソート。
// 詳細画面では特定 exercise の全 example を表示し、採点 usecase では同じ全件を
// テストケースとしてループ実行する。
//
// 全 メソッド は I/O 境界 として `ctx context.Context` を 第 1 引数 で 受ける。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// masterExerciseExampleRepository (GORM)。
type MasterExerciseExampleRepository interface {
	ListByExerciseID(ctx context.Context, exerciseID uint64) ([]domain.MasterExerciseExample, error)
	ListByExerciseIDs(ctx context.Context, exerciseIDs []uint64) (map[uint64][]domain.MasterExerciseExample, error)
}
