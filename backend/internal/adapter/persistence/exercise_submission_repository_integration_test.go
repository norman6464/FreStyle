//go:build integration

package persistence_test

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestExerciseSubmissionRepository_Integration は提出履歴の作成・一覧（絞り込みと並び順）・
// 正解/提出の有無判定を実 Postgres で固定する。
func TestExerciseSubmissionRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewExerciseSubmissionRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "exercise_submissions")

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("Create は採番 ID を書き戻し全フィールドを保存する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "exercise_submissions")
		s := &domain.ExerciseSubmission{
			UserID: 7, ExerciseKind: domain.ExerciseKindMaster, ExerciseID: 100,
			SubmittedCode: "print(1)", Stdout: "1", Stderr: "", ExitCode: 0,
			IsCorrect: true, SubmittedAt: base,
		}
		require.NoError(t, repo.Create(ctx, s))
		require.NotZero(t, s.ID, "採番 ID が書き戻る")

		rows, err := repo.ListByUserAndExercise(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		require.Equal(t, s.ID, rows[0].ID)
		require.Equal(t, "print(1)", rows[0].SubmittedCode)
		require.Equal(t, "1", rows[0].Stdout)
		require.Equal(t, "", rows[0].Stderr, "空文字はそのまま空文字で往復する")
		require.True(t, rows[0].IsCorrect)
		require.Equal(t, base.Unix(), rows[0].SubmittedAt.Unix())
	})

	t.Run("ListByUserAndExercise は user/exercise/kind で絞り submitted_at desc, id desc", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "exercise_submissions")
		// 同一 user/exercise/kind の 3 件。うち 2 件は submitted_at 同着 → id 降順で解決。
		mk := func(kind string, exID uint64, at time.Time, correct bool) *domain.ExerciseSubmission {
			return &domain.ExerciseSubmission{
				UserID: 7, ExerciseKind: kind, ExerciseID: exID,
				SubmittedCode: "x", IsCorrect: correct, SubmittedAt: at,
			}
		}
		older := mk(domain.ExerciseKindMaster, 100, base, false)
		newerA := mk(domain.ExerciseKindMaster, 100, base.Add(time.Hour), false)
		newerB := mk(domain.ExerciseKindMaster, 100, base.Add(time.Hour), true)
		require.NoError(t, repo.Create(ctx, older))
		require.NoError(t, repo.Create(ctx, newerA))
		require.NoError(t, repo.Create(ctx, newerB))
		// 別 kind / 別 exercise / 別 user は混ざらない。
		require.NoError(t, repo.Create(ctx, mk(domain.ExerciseKindCompany, 100, base, true)))
		require.NoError(t, repo.Create(ctx, mk(domain.ExerciseKindMaster, 200, base, true)))
		other := mk(domain.ExerciseKindMaster, 100, base, true)
		other.UserID = 8
		require.NoError(t, repo.Create(ctx, other))

		rows, err := repo.ListByUserAndExercise(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.Len(t, rows, 3, "user7 の master/exercise100 のみ")
		// submitted_at 新しい順、同着は id 降順（newerB は newerA より後に採番）。
		require.Equal(t, newerB.ID, rows[0].ID)
		require.Equal(t, newerA.ID, rows[1].ID)
		require.Equal(t, older.ID, rows[2].ID)
	})

	t.Run("HasSolved は正解提出があるときだけ true", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "exercise_submissions")
		// 不正解だけ → false。
		require.NoError(t, repo.Create(ctx, &domain.ExerciseSubmission{
			UserID: 7, ExerciseKind: domain.ExerciseKindMaster, ExerciseID: 100,
			SubmittedCode: "x", IsCorrect: false, SubmittedAt: base,
		}))
		solved, err := repo.HasSolved(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.False(t, solved)

		attempted, err := repo.HasAttempted(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.True(t, attempted, "提出はあるので attempted は true")

		// 正解を 1 件足すと true。
		require.NoError(t, repo.Create(ctx, &domain.ExerciseSubmission{
			UserID: 7, ExerciseKind: domain.ExerciseKindMaster, ExerciseID: 100,
			SubmittedCode: "x", IsCorrect: true, SubmittedAt: base.Add(time.Minute),
		}))
		solved, err = repo.HasSolved(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.True(t, solved)

		// 別 kind の正解は master の判定に混ざらない。
		require.NoError(t, repo.Create(ctx, &domain.ExerciseSubmission{
			UserID: 7, ExerciseKind: domain.ExerciseKindCompany, ExerciseID: 999,
			SubmittedCode: "x", IsCorrect: true, SubmittedAt: base,
		}))
		solvedOther, err := repo.HasSolved(ctx, 7, 999, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.False(t, solvedOther, "company の正解は master の 999 を解いたことにしない")
	})

	t.Run("HasAttempted は未提出なら false", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "exercise_submissions")
		attempted, err := repo.HasAttempted(ctx, 7, 100, domain.ExerciseKindMaster)
		require.NoError(t, err)
		require.False(t, attempted)
	})
}
