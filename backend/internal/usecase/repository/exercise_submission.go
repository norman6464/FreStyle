package repository

import "github.com/norman6464/FreStyle/backend/internal/domain"

// ExerciseSubmissionRepository は演習提出履歴の永続化を担う。
//
// 履歴は append-only。ユーザ × 問題で複数行を許容し、合否や最新提出時刻を集計する。
// 集計クエリ ( CountUsersSolved / CountTotal ) は exercise_id 単位で
// 一覧ページの「正答者数」「提出数」表示に使う。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// exerciseSubmissionRepository (GORM)。
type ExerciseSubmissionRepository interface {
	// Create は新しい提出を保存する。
	Create(submission *domain.ExerciseSubmission) error

	// ListByUserAndExercise は user × (kind, exercise_id) の履歴を新しい順に返す。
	ListByUserAndExercise(userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error)

	// HasSolved は user が exercise を 1 回でも is_correct=true で解いたかを返す。
	HasSolved(userID, exerciseID uint64, kind string) (bool, error)

	// HasAttempted は user が exercise に対して 1 回でも提出したかを返す。
	HasAttempted(userID, exerciseID uint64, kind string) (bool, error)

	// BatchUserStatuses は user の (kind, exerciseIDs) について、
	//   exercise_id -> "solved" / "in_progress" / ""
	// を返す。一覧ページの N+1 を避ける用途。
	// "" は未提出（map に key が存在しない）を表す。
	BatchUserStatuses(userID uint64, exerciseIDs []uint64, kind string) (map[uint64]string, error)

	// ExerciseStats は exercise_id 単位の集計を返す。
	ExerciseStats(exerciseID uint64, kind string) (ExerciseSubmissionStats, error)

	// ExerciseStatsBatch は複数 exercise_id をまとめて集計する。一覧ページ用。
	ExerciseStatsBatch(exerciseIDs []uint64, kind string) (map[uint64]ExerciseSubmissionStats, error)
}

// ExerciseSubmissionStats は問題単位の集計値。
type ExerciseSubmissionStats struct {
	TotalSubmissions int64 `json:"totalSubmissions"`
	SolvedUsers      int64 `json:"solvedUsers"`
}
