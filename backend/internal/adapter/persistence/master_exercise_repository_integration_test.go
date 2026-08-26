//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// insertMasterExerciseSQL は master_exercises へ 1 行入れる（repository を介さず前提データを置く）。
// 末尾に id 列を足した派生も同じ引数順で使えるよう、列と値をこの順で固定する。
const insertMasterExerciseSQL = `INSERT INTO master_exercises
	(slug, language, sort_order, category, title, description, starter_code,
	 hint_text, expected_output, mode, explanation, difficulty, is_published, chapter_id,
	 created_at, updated_at)
 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now(), now())`

// insertMasterExercise は 1 行入れて、採番された id を ex へ書き戻す（id 指定時はその値を使う）。
// mode / difficulty は列に既定値があり、ゼロ値のまま書くと既定と違う行になるのでここで補う。
func insertMasterExercise(ctx context.Context, t *testing.T, db *sql.DB, ex *domain.MasterExercise) {
	t.Helper()
	mode := ex.Mode
	if mode == "" {
		mode = domain.ExerciseModeExecute
	}
	difficulty := ex.Difficulty
	if difficulty == 0 {
		difficulty = 1
	}
	var chapterID any
	if ex.ChapterID != nil {
		chapterID = int64(*ex.ChapterID)
	}
	args := []any{
		ex.Slug, ex.Language, ex.SortOrder, ex.Category, ex.Title, ex.Description, ex.StarterCode,
		ex.HintText, ex.ExpectedOutput, mode, ex.Explanation, difficulty, ex.IsPublished, chapterID,
	}
	if ex.ID == 0 {
		require.NoError(t, db.QueryRowContext(ctx, insertMasterExerciseSQL+` RETURNING id`, args...).Scan(&ex.ID))
		return
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO master_exercises
			(slug, language, sort_order, category, title, description, starter_code,
			 hint_text, expected_output, mode, explanation, difficulty, is_published, chapter_id,
			 created_at, updated_at, id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, now(), now(), $15)`,
		append(args, ex.ID)...)
	require.NoError(t, err)
}

// TestMasterExerciseRepository_ListWithStatusByLanguage_Integration は、3 クエリを 1 本の
// LEFT JOIN + FILTER に統合した一覧クエリを実 Postgres で検証する。
func TestMasterExerciseRepository_ListWithStatusByLanguage_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	exRepo := persistence.NewMasterExerciseRepository(sqlDB)
	subRepo := persistence.NewExerciseSubmissionRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "master_exercises", "exercise_submissions")

	// 問題: php-1(公開) / php-2(公開) / go-1(公開) / draft-1(非公開) を用意。
	exercises := []domain.MasterExercise{
		{Slug: "php-1", Language: "php", Title: "PHP1", SortOrder: 1, IsPublished: true},
		{Slug: "php-2", Language: "php", Title: "PHP2", SortOrder: 2, IsPublished: true},
		{Slug: "go-1", Language: "go", Title: "Go1", SortOrder: 3, IsPublished: true},
		{Slug: "draft-1", Language: "php", Title: "Draft", SortOrder: 4, IsPublished: false},
	}
	for i := range exercises {
		insertMasterExercise(ctx, t, sqlDB, &exercises[i])
	}
	// is_published は列の既定値が true。draft-1 を明示的に非公開へ更新して非公開除外を検証する。
	_, err := sqlDB.ExecContext(ctx,
		`UPDATE master_exercises SET is_published = $1, updated_at = now() WHERE slug = $2`,
		false, "draft-1")
	require.NoError(t, err)
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
// is_published は列の既定値が true なので、非公開にしたい行は INSERT 後に明示 UPDATE する。
func seedMasterExercisesForContract(t *testing.T, db *sql.DB, ctx context.Context) map[string]uint64 {
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
		insertMasterExercise(ctx, t, db, &rows[i])
		ids[rows[i].Slug] = rows[i].ID
	}
	_, err := db.ExecContext(ctx,
		`UPDATE master_exercises SET is_published = $1, updated_at = now() WHERE slug = $2`,
		false, "php-draft")
	require.NoError(t, err)
	return ids
}

// TestMasterExerciseRepository_ReadContract_Integration は単純取得系の契約
// （言語フィルタ / 非公開除外 / 並び順とタイブレーク / 未存在の not found）を実 Postgres で固定する。
func TestMasterExerciseRepository_ReadContract_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewMasterExerciseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "master_exercises", "exercise_submissions")
	ids := seedMasterExercisesForContract(t, sqlDB, ctx)

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
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewMasterExerciseRepository(sqlDB)
	subRepo := persistence.NewExerciseSubmissionRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "master_exercises", "exercise_submissions")
	ids := seedMasterExercisesForContract(t, sqlDB, ctx)

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
