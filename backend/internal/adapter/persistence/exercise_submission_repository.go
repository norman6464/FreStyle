package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// exerciseSubmissionRepository は [repository.ExerciseSubmissionRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type exerciseSubmissionRepository struct {
	db *sql.DB
}

func NewExerciseSubmissionRepository(db *sql.DB) repository.ExerciseSubmissionRepository {
	return &exerciseSubmissionRepository{db: db}
}

func toDomainExerciseSubmission(row sqlcgen.ExerciseSubmission) domain.ExerciseSubmission {
	return domain.ExerciseSubmission{
		ID:            uint64(row.ID),
		UserID:        uint64(row.UserID),
		ExerciseKind:  row.ExerciseKind,
		ExerciseID:    uint64(row.ExerciseID),
		SubmittedCode: row.SubmittedCode,
		Stdout:        row.Stdout.String, // NULL は "" に倒す（GORM の非ポインタ string と同じ）
		Stderr:        row.Stderr.String,
		ExitCode:      int(row.ExitCode),
		IsCorrect:     row.IsCorrect,
		SubmittedAt:   row.SubmittedAt,
	}
}

func (r *exerciseSubmissionRepository) Create(ctx context.Context, submission *domain.ExerciseSubmission) error {
	uid, ok := toInt64ID(submission.UserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("user_id", submission.UserID)
	}
	exID, ok := toInt64ID(submission.ExerciseID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("exercise_id", submission.ExerciseID)
	}
	id, err := sqlcgen.New(r.db).InsertExerciseSubmission(ctx, sqlcgen.InsertExerciseSubmissionParams{
		UserID:        uid,
		ExerciseKind:  submission.ExerciseKind,
		ExerciseID:    exID,
		SubmittedCode: submission.SubmittedCode,
		// GORM 版は非ポインタ string を常に書いていたので空文字も '' として保存する。
		Stdout:      sql.NullString{String: submission.Stdout, Valid: true},
		Stderr:      sql.NullString{String: submission.Stderr, Valid: true},
		ExitCode:    int64(submission.ExitCode),
		IsCorrect:   submission.IsCorrect,
		SubmittedAt: submission.SubmittedAt,
	})
	if err != nil {
		return err
	}
	submission.ID = uint64(id)
	return nil
}

func (r *exerciseSubmissionRepository) ListByUserAndExercise(ctx context.Context, userID, exerciseID uint64, kind string) ([]domain.ExerciseSubmission, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.ExerciseSubmission{}, nil
	}
	exID, ok := toInt64ID(exerciseID)
	if !ok {
		return []domain.ExerciseSubmission{}, nil
	}
	rows, err := sqlcgen.New(r.db).ListSubmissionsByUserAndExercise(ctx, sqlcgen.ListSubmissionsByUserAndExerciseParams{
		UserID:       uid,
		ExerciseID:   exID,
		ExerciseKind: kind,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ExerciseSubmission, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainExerciseSubmission(row))
	}
	return out, nil
}

func (r *exerciseSubmissionRepository) HasSolved(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return false, nil
	}
	exID, ok := toInt64ID(exerciseID)
	if !ok {
		return false, nil
	}
	return sqlcgen.New(r.db).ExistsCorrectSubmission(ctx, sqlcgen.ExistsCorrectSubmissionParams{
		UserID:       uid,
		ExerciseID:   exID,
		ExerciseKind: kind,
	})
}

func (r *exerciseSubmissionRepository) HasAttempted(ctx context.Context, userID, exerciseID uint64, kind string) (bool, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return false, nil
	}
	exID, ok := toInt64ID(exerciseID)
	if !ok {
		return false, nil
	}
	return sqlcgen.New(r.db).ExistsSubmission(ctx, sqlcgen.ExistsSubmissionParams{
		UserID:       uid,
		ExerciseID:   exID,
		ExerciseKind: kind,
	})
}
