package usecase

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// ListMasterExercisesUseCase は指定言語の運営マスタ演習問題一覧を返す。
//
// 旧 ListPhpExercisesUseCase の汎用版。言語固定の API（例: /php/exercises）を
// 残しつつ、本 usecase は language を引数で受け取って多言語対応の準備をする。
type ListMasterExercisesUseCase struct {
	repo repository.MasterExerciseRepository
}

func NewListMasterExercisesUseCase(repo repository.MasterExerciseRepository) *ListMasterExercisesUseCase {
	return &ListMasterExercisesUseCase{repo: repo}
}

// Execute は language 指定があれば該当言語のみ、空文字なら全言語の問題を返す。
func (uc *ListMasterExercisesUseCase) Execute(language string) ([]domain.MasterExercise, error) {
	return uc.repo.ListByLanguage(language)
}
