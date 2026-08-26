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
	"gorm.io/gorm"
)

// TestMasterExerciseRepository_ListWithStatusByLanguage_Integration は、3 クエリを 1 本の
// LEFT JOIN + FILTER に統合した一覧クエリを実 Postgres で検証する。
func TestMasterExerciseRepository_ListWithStatusByLanguage_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	exRepo := persistence.NewMasterExerciseRepository(sqlDB)
	subRepo := persistence.NewExerciseSubmissionRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercises", "exercise_submissions")

	// 問題: php-1(公開) / php-2(公開) / go-1(公開) / draft-1(非公開) を用意。
	exercises := []domain.MasterExercise{
		{Slug: "php-1", Language: "php", Title: "PHP1", SortOrder: 1, IsPublished: true},
		{Slug: "php-2", Language: "php", Title: "PHP2", SortOrder: 2, IsPublished: true},
		{Slug: "go-1", Language: "go", Title: "Go1", SortOrder: 3, IsPublished: true},
		{Slug: "draft-1", Language: "php", Title: "Draft", SortOrder: 4, IsPublished: false},
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

	t.Run("言語フィルタ + 非公開除外 + sort_order 昇順", func(t *testing.T) {
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

// seedMasterExercisesForContract は単純取得系（ListByLanguage / GetByID / GetBySlug /
// SummaryByLanguage）の検証データを用意し、slug → id の対応を返す。
// is_published は GORM タグ `default:true` のため bool ゼロ値では非公開にならない。
// 非公開にしたい行は Create 後に明示 UPDATE する（既存テストと同じ手当て）。
func seedMasterExercisesForContract(t *testing.T, db *gorm.DB, ctx context.Context) map[string]uint64 {
	t.Helper()
	rows := []domain.MasterExercise{
		// sort_order 同値（10）で 2 件置き、id 昇順のタイブレークが効くことを見る。
		{Slug: "php-b", Language: "php", Title: "PHP B", SortOrder: 10, IsPublished: true, Difficulty: 2, Mode: domain.ExerciseModeExecute},
		{Slug: "php-a", Language: "php", Title: "PHP A", SortOrder: 10, IsPublished: true, Difficulty: 3, Mode: domain.ExerciseModeQA},
		{Slug: "php-z", Language: "php", Title: "PHP Z", SortOrder: 5, IsPublished: true},
		{Slug: "go-a", Language: "go", Title: "Go A", SortOrder: 1, IsPublished: true},
		{Slug: "php-draft", Language: "php", Title: "PHP Draft", SortOrder: 1, IsPublished: false},
	}
	ids := make(map[string]uint64, len(rows))
	for i := range rows {
		require.NoError(t, db.WithContext(ctx).Create(&rows[i]).Error)
		ids[rows[i].Slug] = rows[i].ID
	}
	require.NoError(t, db.WithContext(ctx).Model(&domain.MasterExercise{}).
		Where("slug = ?", "php-draft").Update("is_published", false).Error)
	return ids
}

// TestMasterExerciseRepository_ReadContract_Integration は単純取得系の契約
// （言語フィルタ / 非公開除外 / 並び順とタイブレーク / 未存在の not found）を実 Postgres で固定する。
func TestMasterExerciseRepository_ReadContract_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewMasterExerciseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercises", "exercise_submissions")
	ids := seedMasterExercisesForContract(t, db, ctx)

	t.Run("ListByLanguage は言語で絞り非公開を除く", func(t *testing.T) {
		rows, err := repo.ListByLanguage(ctx, "php")
		require.NoError(t, err)
		slugs := make([]string, 0, len(rows))
		for _, r := range rows {
			slugs = append(slugs, r.Slug)
		}
		// sort_order 昇順 → 同値(10)は id 昇順で php-b(先に作成) → php-a。
		require.Equal(t, []string{"php-z", "php-b", "php-a"}, slugs)
	})

	t.Run("ListByLanguage(空文字)は全言語の公開分", func(t *testing.T) {
		rows, err := repo.ListByLanguage(ctx, "")
		require.NoError(t, err)
		slugs := make([]string, 0, len(rows))
		for _, r := range rows {
			slugs = append(slugs, r.Slug)
		}
		// 言語をまたいで sort_order 昇順 → 同値は id 昇順。
		require.Equal(t, []string{"go-a", "php-z", "php-b", "php-a"}, slugs)
	})

	t.Run("ListByLanguage は該当なしで空スライス（nil ではない）", func(t *testing.T) {
		rows, err := repo.ListByLanguage(ctx, "no-such-language")
		require.NoError(t, err)
		require.NotNil(t, rows)
		require.Empty(t, rows)
	})

	t.Run("GetByID は全列を詰めて返す", func(t *testing.T) {
		got, err := repo.GetByID(ctx, ids["php-a"])
		require.NoError(t, err)
		require.Equal(t, "php-a", got.Slug)
		require.Equal(t, "php", got.Language)
		require.Equal(t, "PHP A", got.Title)
		require.Equal(t, 10, got.SortOrder)
		require.Equal(t, int16(3), got.Difficulty)
		require.Equal(t, domain.ExerciseModeQA, got.Mode)
		require.True(t, got.IsPublished)
		require.Nil(t, got.ChapterID)
		require.False(t, got.CreatedAt.IsZero())
		require.False(t, got.UpdatedAt.IsZero())
	})

	t.Run("GetByID は非公開でも取得できる", func(t *testing.T) {
		got, err := repo.GetByID(ctx, ids["php-draft"])
		require.NoError(t, err)
		require.False(t, got.IsPublished)
	})

	t.Run("GetByID の未存在は domain.ErrNotFound", func(t *testing.T) {
		_, err := repo.GetByID(ctx, noSuchID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("GetBySlug は slug で 1 件返す", func(t *testing.T) {
		got, err := repo.GetBySlug(ctx, "go-a")
		require.NoError(t, err)
		require.Equal(t, ids["go-a"], got.ID)
		require.Equal(t, "go", got.Language)
	})

	t.Run("GetBySlug の未存在は domain.ErrNotFound", func(t *testing.T) {
		_, err := repo.GetBySlug(ctx, "no-such-slug")
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

// TestMasterExerciseRepository_SummaryByLanguage_Integration は言語別集計の契約
// （公開分だけ数える / 正解済みは問題単位で 1 / 他人の提出は数えない / 言語昇順）を固定する。
func TestMasterExerciseRepository_SummaryByLanguage_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewMasterExerciseRepository(sqlDB)
	subRepo := persistence.NewExerciseSubmissionRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercises", "exercise_submissions")
	ids := seedMasterExercisesForContract(t, db, ctx)

	// user7: php-a を 2 回正解（問題単位では 1 件）+ php-b を不正解のみ。
	// user8: php-z を正解（user7 の集計には出ない）。
	// company 種別の提出は master の集計に混ぜない。
	subs := []domain.ExerciseSubmission{
		{UserID: 7, ExerciseID: ids["php-a"], ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
		{UserID: 7, ExerciseID: ids["php-a"], ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
		{UserID: 7, ExerciseID: ids["php-b"], ExerciseKind: domain.ExerciseKindMaster, IsCorrect: false},
		{UserID: 8, ExerciseID: ids["php-z"], ExerciseKind: domain.ExerciseKindMaster, IsCorrect: true},
		{UserID: 7, ExerciseID: ids["go-a"], ExerciseKind: domain.ExerciseKindCompany, IsCorrect: true},
	}
	for i := range subs {
		require.NoError(t, subRepo.Create(ctx, &subs[i]))
	}

	t.Run("ログインユーザの正解数が言語ごとに付く", func(t *testing.T) {
		rows, err := repo.SummaryByLanguage(ctx, 7)
		require.NoError(t, err)
		require.Equal(t, []repository.ExerciseLanguageSummary{
			// language 昇順。go は公開 1 件・user7 の master 正解なし（company 種別は数えない）。
			{Language: "go", Total: 1, Solved: 0},
			// php は公開 3 件（draft 除外）。user7 の正解は php-a のみ。
			{Language: "php", Total: 3, Solved: 1},
		}, rows)
	})

	t.Run("未ログイン(userID=0)は solved が全て 0", func(t *testing.T) {
		rows, err := repo.SummaryByLanguage(ctx, 0)
		require.NoError(t, err)
		require.Equal(t, []repository.ExerciseLanguageSummary{
			{Language: "go", Total: 1, Solved: 0},
			{Language: "php", Total: 3, Solved: 0},
		}, rows)
	})
}
