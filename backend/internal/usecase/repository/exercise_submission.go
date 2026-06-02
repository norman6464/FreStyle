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

	// BatchUserStatuses は exercise_id -> "solved" / "in_progress" を返す（未提出は key なし、N+1 回避）。
	BatchUserStatuses(ctx context.Context, userID uint64, exerciseIDs []uint64, kind string) (map[uint64]string, error)

	// ExerciseStats は exercise_id 単位の集計を返す。
	ExerciseStats(ctx context.Context, exerciseID uint64, kind string) (ExerciseSubmissionStats, error)

	// ExerciseStatsBatch は複数 exercise_id をまとめて集計する。
	ExerciseStatsBatch(ctx context.Context, exerciseIDs []uint64, kind string) (map[uint64]ExerciseSubmissionStats, error)
}

// ExerciseSubmissionStats は問題単位の集計値。
type ExerciseSubmissionStats struct {
	TotalSubmissions int64 `json:"totalSubmissions"`
	SolvedUsers      int64 `json:"solvedUsers"`
}
