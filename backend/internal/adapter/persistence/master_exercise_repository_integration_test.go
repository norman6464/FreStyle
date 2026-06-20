//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestMasterExerciseRepository_ListWithStatusByLanguage_Integration は、3 クエリを 1 本の
// LEFT JOIN + FILTER に統合した一覧クエリを実 Postgres で検証する。
func TestMasterExerciseRepository_ListWithStatusByLanguage_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	exRepo := persistence.NewMasterExerciseRepository(db)
	subRepo := persistence.NewExerciseSubmissionRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercises", "exercise_submissions")

	// 問題: php-1(公開) / php-2(公開) / go-1(公開) / draft-1(非公開) を用意。
	exercises := []domain.MasterExercise{
		{Slug: "php-1", Language: "php", Title: "PHP1", OrderIndex: 1, IsPublished: true},
		{Slug: "php-2", Language: "php", Title: "PHP2", OrderIndex: 2, IsPublished: true},
		{Slug: "go-1", Language: "go", Title: "Go1", OrderIndex: 3, IsPublished: true},
		{Slug: "draft-1", Language: "php", Title: "Draft", OrderIndex: 4, IsPublished: false},
	}
	for i := range exercises {
		require.NoError(t, db.WithContext(ctx).Create(&exercises[i]).Error)
	}
	// is_published は GORM タグ `default:true` のため、bool ゼロ値 (false) を Create に渡しても
	// 「未指定」とみなされ DB 側で true になる。draft-1 を明示的に非公開へ更新して非公開除外を検証する。
	require.NoError(t, db.WithContext(ctx).Model(&domain.MasterExercise{}).
		Where("slug = ?", "draft-1").Update("is_published", false).Error)
	php1ID := exercises[0].ID

	// 提出: php-1 に user7 正解 + user7 不正解 + user8 正解（総提出3 / 正解 distinct 2）。
	subs := []domain.ExerciseSubmission{
		{UserID: 7, ExerciseID: php1ID, ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
		{UserID: 7, ExerciseID: php1ID, ExerciseKind: domain.ExerciseKindMaster, IsCorrect: false},
		{UserID: 8, ExerciseID: php1ID, ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
	}
	for i := range subs {
		require.NoError(t, subRepo.Create(ctx, &subs[i]))
	}

	in := func(userID uint64, language string) repository.ListWithStatusInput {
		return repository.ListWithStatusInput{UserID: userID, Language: language}
	}

	t.Run("言語フィルタ + 非公開除外 + order_index 昇順", func(t *testing.T) {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, in(0, "php"))
		require.NoError(t, err)
		require.Len(t, rows, 2, "php の公開問題のみ（draft は除外）")
		require.Equal(t, "php-1", rows[0].Slug)
		require.Equal(t, "php-2", rows[1].Slug)
	})

	t.Run("未ログイン(userID=0)は status 空 + 全体集計は付く", func(t *testing.T) {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, in(0, "php"))
		require.NoError(t, err)
		require.Equal(t, "", rows[0].Status)
		require.Equal(t, int64(3), rows[0].Stats.TotalSubmissions)
		require.Equal(t, int64(2), rows[0].Stats.SolvedUsers)
		require.Equal(t, int64(0), rows[1].Stats.TotalSubmissions)
	})

	t.Run("ログインユーザの status(solved)が付く", func(t *testing.T) {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, in(7, "php"))
		require.NoError(t, err)
		require.Equal(t, "solved", rows[0].Status, "user7 は php-1 を正解済み")
		require.Equal(t, "", rows[1].Status, "php-2 は未提出")
	})

	t.Run("language 空なら全公開問題", func(t *testing.T) {
		rows, err := exRepo.ListWithStatusByLanguage(ctx, in(0, ""))
		require.NoError(t, err)
		require.Len(t, rows, 3, "php-1 / php-2 / go-1（draft 除外）")
	})

	t.Run("LIMIT/OFFSET でページネーション", func(t *testing.T) {
		rows1, err := exRepo.ListWithStatusByLanguage(ctx, repository.ListWithStatusInput{UserID: 0, Language: "", Offset: 0, Limit: 2})
		require.NoError(t, err)
		require.Len(t, rows1, 2)

		rows2, err := exRepo.ListWithStatusByLanguage(ctx, repository.ListWithStatusInput{UserID: 0, Language: "", Offset: 2, Limit: 2})
		require.NoError(t, err)
		require.Len(t, rows2, 1, "3 件目（残り 1 件）")

		require.NotEqual(t, rows1[0].Slug, rows2[0].Slug, "ページが重複しない")
	})
}
