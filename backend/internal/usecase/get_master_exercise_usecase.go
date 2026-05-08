package usecase

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// GetMasterExerciseUseCase は指定 ID の運営マスタ演習問題を返す。
//
// 旧 GetPhpExerciseUseCase の汎用版。ID は MasterExercise の primary key（uint64）。
type GetMasterExerciseUseCase struct {
	repo repository.MasterExerciseRepository
}

func NewGetMasterExerciseUseCase(repo repository.MasterExerciseRepository) *GetMasterExerciseUseCase {
	return &GetMasterExerciseUseCase{repo: repo}
}

func (uc *GetMasterExerciseUseCase) Execute(id uint64) (*domain.MasterExercise, error) {
	return uc.repo.GetByID(id)
}
