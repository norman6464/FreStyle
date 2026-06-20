package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// masterExerciseRepository は [repository.MasterExerciseRepository] の GORM 実装。
type masterExerciseRepository struct {
	db *gorm.DB
}

func NewMasterExerciseRepository(db *gorm.DB) repository.MasterExerciseRepository {
	return &masterExerciseRepository{db: db}
}

func (r *masterExerciseRepository) ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error) {
	var exercises []domain.MasterExercise
	q := r.db.WithContext(ctx).Where("is_published = ?", true)
	if language != "" {
		q = q.Where("language = ?", language)
	}
	if err := q.Order("order_index asc").Find(&exercises).Error; err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *masterExerciseRepository) GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.WithContext(ctx).First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *masterExerciseRepository) GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&exercise).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

// ListWithStatusByLanguage は公開済み問題を「current user の提出状態 + 全体集計」付きで 1 クエリで返す。
// in.Limit > 0 のとき LIMIT/OFFSET を適用する。Limit=0 は全件取得。
// userID=0（未ログイン）は usr サブクエリが空になり status は全て "".
func (r *masterExerciseRepository) ListWithStatusByLanguage(ctx context.Context, in repository.ListWithStatusInput) ([]repository.MasterExerciseWithStatus, error) {
	type row struct {
		domain.MasterExercise

		UserStatus       string `gorm:"column:user_status"`
		TotalSubmissions int64  `gorm:"column:total_submissions"`
		SolvedUsers      int64  `gorm:"column:solved_users"`
	}

	const baseQ = `
SELECT e.*,
       COALESCE(agg.total_submissions, 0) AS total_submissions,
       COALESCE(agg.solved_users, 0)      AS solved_users,
       CASE
           WHEN usr.any_solved IS TRUE  THEN 'solved'
           WHEN usr.any_solved IS FALSE THEN 'in_progress'
           ELSE ''
       END AS user_status
FROM master_exercises e
LEFT JOIN (
    SELECT exercise_id,
           COUNT(*)                                          AS total_submissions,
           COUNT(DISTINCT user_id) FILTER (WHERE is_correct) AS solved_users
    FROM exercise_submissions
    WHERE exercise_kind = ?
    GROUP BY exercise_id
) agg ON agg.exercise_id = e.id
LEFT JOIN (
    SELECT exercise_id, BOOL_OR(is_correct) AS any_solved
    FROM exercise_submissions
    WHERE exercise_kind = ? AND user_id = ?
    GROUP BY exercise_id
) usr ON usr.exercise_id = e.id
WHERE e.is_published = TRUE
  AND (? = '' OR e.language = ?)
ORDER BY e.order_index ASC`

	var q string
	var args []interface{}
	args = []interface{}{domain.ExerciseKindMaster, domain.ExerciseKindMaster, in.UserID, in.Language, in.Language}

	if in.Limit > 0 {
		q = baseQ + "\nLIMIT ? OFFSET ?"
		args = append(args, in.Limit, in.Offset)
	} else {
		q = baseQ
	}

	var rows []row
	if err := r.db.WithContext(ctx).Raw(q, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]repository.MasterExerciseWithStatus, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.MasterExerciseWithStatus{
			MasterExercise: row.MasterExercise,
			Status:         row.UserStatus,
			Stats: repository.ExerciseSubmissionStats{
				TotalSubmissions: row.TotalSubmissions,
				SolvedUsers:      row.SolvedUsers,
			},
		})
	}
	return out, nil
}
