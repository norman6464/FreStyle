package usecase

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
)

// MasterExerciseWithStatus は一覧ページに渡す「問題 + 現在ユーザの状態 + 全体集計」のセット。
//
// Status の値:
//   - "solved"      … current user が 1 度でも is_correct=true で解いた
//   - "in_progress" … current user が提出はあるが正答していない
//   - ""            … current user が未提出（未着手）
//
// Stats は全ユーザ合計の集計値。
type MasterExerciseWithStatus struct {
	domain.MasterExercise

	Status string                                   `json:"status"`
	Stats  legacyrepository.ExerciseSubmissionStats `json:"stats"`
}

// ListMasterExercisesWithStatusUseCase は問題一覧 + 各問題の current user 状態 + 集計を返す。
//
// N+1 を避けるため、 user status と stats はそれぞれ batch クエリで取得する。
type ListMasterExercisesWithStatusUseCase struct {
	exercises   legacyrepository.MasterExerciseRepository
	submissions legacyrepository.ExerciseSubmissionRepository
}

func NewListMasterExercisesWithStatusUseCase(
	exercises legacyrepository.MasterExerciseRepository,
	submissions legacyrepository.ExerciseSubmissionRepository,
) *ListMasterExercisesWithStatusUseCase {
	return &ListMasterExercisesWithStatusUseCase{exercises: exercises, submissions: submissions}
}

// ListMasterExercisesWithStatusInput は入力。 UserID=0 は未ログイン扱いで status は全部 ""。
type ListMasterExercisesWithStatusInput struct {
	UserID   uint64
	Language string
}

func (uc *ListMasterExercisesWithStatusUseCase) Execute(in ListMasterExercisesWithStatusInput) ([]MasterExerciseWithStatus, error) {
	exercises, err := uc.exercises.ListByLanguage(in.Language)
	if err != nil {
		return nil, err
	}
	if len(exercises) == 0 {
		return []MasterExerciseWithStatus{}, nil
	}
	ids := make([]uint64, 0, len(exercises))
	for _, e := range exercises {
		ids = append(ids, e.ID)
	}

	statusMap := make(map[uint64]string)
	if in.UserID != 0 {
		statusMap, err = uc.submissions.BatchUserStatuses(in.UserID, ids, domain.ExerciseKindMaster)
		if err != nil {
			return nil, err
		}
	}
	statsMap, err := uc.submissions.ExerciseStatsBatch(ids, domain.ExerciseKindMaster)
	if err != nil {
		return nil, err
	}

	out := make([]MasterExerciseWithStatus, 0, len(exercises))
	for _, e := range exercises {
		out = append(out, MasterExerciseWithStatus{
			MasterExercise: e,
			Status:         statusMap[e.ID],
			Stats:          statsMap[e.ID],
		})
	}
	return out, nil
}
