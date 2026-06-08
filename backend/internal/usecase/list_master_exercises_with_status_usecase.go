package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// MasterExerciseWithStatus は read model（repository 定義）を handler / OpenAPI 向けに再エクスポートした別名。
// 正準型は repository パッケージにあり、persistence はそちらを返す（層境界のため）。
type MasterExerciseWithStatus = repository.MasterExerciseWithStatus

// ListMasterExercisesWithStatusUseCase は問題一覧 + 各問題の current user 状態 + 集計を返す。
// 取得は repository が 1 クエリ（master_exercises ⟕ exercise_submissions 集計）で行い、N+1 / 多段往復を避ける。
type ListMasterExercisesWithStatusUseCase struct {
	exercises repository.MasterExerciseRepository
}

func NewListMasterExercisesWithStatusUseCase(
	exercises repository.MasterExerciseRepository,
) *ListMasterExercisesWithStatusUseCase {
	return &ListMasterExercisesWithStatusUseCase{exercises: exercises}
}

// ListMasterExercisesWithStatusInput は入力。 UserID=0 は未ログイン扱いで status は全部 ""。
type ListMasterExercisesWithStatusInput struct {
	UserID   uint64
	Language string
}

func (uc *ListMasterExercisesWithStatusUseCase) Execute(ctx context.Context, in ListMasterExercisesWithStatusInput) ([]repository.MasterExerciseWithStatus, error) {
	return uc.exercises.ListWithStatusByLanguage(ctx, in.UserID, in.Language)
}
