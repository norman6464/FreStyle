package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ExerciseSubmissionRepository は演習提出履歴（append-only）の永続化と集計を担う。
type ExerciseSubmissionRepository interface {
	Create(ctx context.Context, submission *domain.ExerciseSubmission) error

	// ListByUserAndExercise は user × (kind, exercise_id) の履歴を新しい順に返す。
	ListByUserAndExercise(ctx context.Context, userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error)

	// HasSolved は user が exercise を 1 回でも is_correct=true で解いたかを返す。
	HasSolved(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error)

	// HasAttempted は user が exercise に 1 回でも提出したかを返す。
	HasAttempted(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error)
}

// ExerciseSubmissionStats は問題単位の集計値。
type ExerciseSubmissionStats struct {
	TotalSubmissions int64 `json:"totalSubmissions"`
	SolvedUsers      int64 `json:"solvedUsers"`
}
