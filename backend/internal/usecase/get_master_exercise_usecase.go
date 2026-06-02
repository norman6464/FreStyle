package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetMasterExerciseDetailOutput は詳細ページに渡す問題本体 + 入出力例セット。
type GetMasterExerciseDetailOutput struct {
	Exercise *domain.MasterExercise         `json:"exercise"`
	Examples []domain.MasterExerciseExample `json:"examples"`
}

// GetMasterExerciseUseCase は slug 指定で運営マスタ演習 + 入出力例を返す（ID 指定は Execute で互換維持）。
type GetMasterExerciseUseCase struct {
	repo     repository.MasterExerciseRepository
	examples repository.MasterExerciseExampleRepository
}

func NewGetMasterExerciseUseCase(
	repo repository.MasterExerciseRepository,
	examples repository.MasterExerciseExampleRepository,
) *GetMasterExerciseUseCase {
	return &GetMasterExerciseUseCase{repo: repo, examples: examples}
}

// Execute は ID 指定で取得する旧 API 互換。 examples は付かない。
func (uc *GetMasterExerciseUseCase) Execute(ctx context.Context, id uint64) (*domain.MasterExercise, error) {
	return uc.repo.GetByID(ctx, id)
}

// ExecuteBySlug は slug ベースの詳細ページ向けに examples を含めて返す。
// NotFound は handler で 404 に分岐できるようそのまま伝搬し、ex == nil は defensive に弾く。
func (uc *GetMasterExerciseUseCase) ExecuteBySlug(ctx context.Context, slug string) (*GetMasterExerciseDetailOutput, error) {
	ex, err := uc.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", slug)
	}
	examples, err := uc.examples.ListByExerciseID(ctx, ex.ID)
	if err != nil {
		return nil, err
	}
	return &GetMasterExerciseDetailOutput{Exercise: ex, Examples: examples}, nil
}
