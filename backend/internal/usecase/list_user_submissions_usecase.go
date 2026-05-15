package usecase

import (
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListUserMasterSubmissionsInput は履歴一覧 API への入力。
type ListUserMasterSubmissionsInput struct {
	UserID uint64
	Slug   string
}

// ListUserMasterSubmissionsUseCase は current user の指定問題に対する提出履歴を新しい順で返す。
type ListUserMasterSubmissionsUseCase struct {
	exercises   repository.MasterExerciseRepository
	submissions repository.ExerciseSubmissionRepository
}

func NewListUserMasterSubmissionsUseCase(
	exercises repository.MasterExerciseRepository,
	submissions repository.ExerciseSubmissionRepository,
) *ListUserMasterSubmissionsUseCase {
	return &ListUserMasterSubmissionsUseCase{exercises: exercises, submissions: submissions}
}

func (uc *ListUserMasterSubmissionsUseCase) Execute(in ListUserMasterSubmissionsInput) ([]domain.ExerciseSubmission, error) {
	ex, err := uc.exercises.GetBySlug(in.Slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", in.Slug)
	}
	return uc.submissions.ListByUserAndExercise(in.UserID, ex.ID, domain.ExerciseKindMaster)
}
