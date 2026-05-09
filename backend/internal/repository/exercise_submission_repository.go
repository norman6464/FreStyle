package repository

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// ExerciseSubmissionRepository は演習提出履歴の永続化を担う。
//
// 履歴は append-only。ユーザ × 問題で複数行を許容し、合否や最新提出時刻を集計する。
// 集計クエリ ( CountUsersSolved / CountTotal ) は exercise_id 単位で
// 一覧ページの「正答者数」「提出数」表示に使う。
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

type exerciseSubmissionRepository struct {
	db *gorm.DB
}

func NewExerciseSubmissionRepository(db *gorm.DB) ExerciseSubmissionRepository {
	return &exerciseSubmissionRepository{db: db}
}

func (r *exerciseSubmissionRepository) Create(submission *domain.ExerciseSubmission) error {
	return r.db.Create(submission).Error
}

func (r *exerciseSubmissionRepository) ListByUserAndExercise(userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error) {
	var rows []domain.ExerciseSubmission
	if err := r.db.
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ?", userID, exerciseID, kind).
		Order("submitted_at desc, id desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *exerciseSubmissionRepository) HasSolved(userID, exerciseID uint64, kind string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ? AND is_correct = ?", userID, exerciseID, kind, true).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *exerciseSubmissionRepository) HasAttempted(userID, exerciseID uint64, kind string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Where("user_id = ? AND exercise_id = ? AND exercise_kind = ?", userID, exerciseID, kind).
		Limit(1).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// BatchUserStatuses は 1 クエリで user × exerciseIDs の (any submission, any solved) を取り、
// exercise_id -> "solved" / "in_progress" を返す。 結果に含まれない exercise_id は未着手扱い。
func (r *exerciseSubmissionRepository) BatchUserStatuses(userID uint64, exerciseIDs []uint64, kind string) (map[uint64]string, error) {
	result := make(map[uint64]string, len(exerciseIDs))
	if len(exerciseIDs) == 0 {
		return result, nil
	}
	type row struct {
		ExerciseID uint64
		AnySolved  bool
	}
	var rows []row
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Select("exercise_id, BOOL_OR(is_correct) AS any_solved").
		Where("user_id = ? AND exercise_kind = ? AND exercise_id IN ?", userID, kind, exerciseIDs).
		Group("exercise_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		if r.AnySolved {
			result[r.ExerciseID] = "solved"
		} else {
			result[r.ExerciseID] = "in_progress"
		}
	}
	return result, nil
}

func (r *exerciseSubmissionRepository) ExerciseStats(exerciseID uint64, kind string) (ExerciseSubmissionStats, error) {
	var stats ExerciseSubmissionStats
	// COUNT(*) は全提出数。
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Where("exercise_id = ? AND exercise_kind = ?", exerciseID, kind).
		Count(&stats.TotalSubmissions).Error; err != nil {
		return stats, err
	}
	// COUNT(DISTINCT user_id) は正答ユーザ数。
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Where("exercise_id = ? AND exercise_kind = ? AND is_correct = ?", exerciseID, kind, true).
		Distinct("user_id").
		Count(&stats.SolvedUsers).Error; err != nil {
		return stats, err
	}
	return stats, nil
}

// ExerciseStatsBatch は 2 つの GROUP BY クエリで全件まとめて集計する。
func (r *exerciseSubmissionRepository) ExerciseStatsBatch(exerciseIDs []uint64, kind string) (map[uint64]ExerciseSubmissionStats, error) {
	result := make(map[uint64]ExerciseSubmissionStats, len(exerciseIDs))
	if len(exerciseIDs) == 0 {
		return result, nil
	}
	// 提出数の集計。
	type totalRow struct {
		ExerciseID uint64
		Total      int64
	}
	var totals []totalRow
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Select("exercise_id, COUNT(*) AS total").
		Where("exercise_kind = ? AND exercise_id IN ?", kind, exerciseIDs).
		Group("exercise_id").
		Scan(&totals).Error; err != nil {
		return nil, err
	}
	for _, t := range totals {
		s := result[t.ExerciseID]
		s.TotalSubmissions = t.Total
		result[t.ExerciseID] = s
	}
	// 正答ユーザ数の集計（DISTINCT user_id）。
	type solvedRow struct {
		ExerciseID  uint64
		SolvedUsers int64
	}
	var solveds []solvedRow
	if err := r.db.Model(&domain.ExerciseSubmission{}).
		Select("exercise_id, COUNT(DISTINCT user_id) AS solved_users").
		Where("exercise_kind = ? AND exercise_id IN ? AND is_correct = ?", kind, exerciseIDs, true).
		Group("exercise_id").
		Scan(&solveds).Error; err != nil {
		return nil, err
	}
	for _, s := range solveds {
		st := result[s.ExerciseID]
		st.SolvedUsers = s.SolvedUsers
		result[s.ExerciseID] = st
	}
	return result, nil
}
