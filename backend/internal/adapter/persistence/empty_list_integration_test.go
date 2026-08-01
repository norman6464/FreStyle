//go:build integration

package persistence_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 存在しない ID。どのテーブルにも該当行が無い状態を作るために使う。
const noSuchID uint64 = 999_999_999

// assertEmptyJSONArray は「0 件の結果が nil ではなく空スライスで、JSON で [] になる」ことを検証する。
//
// nil スライスは encoding/json で null になり、フロントの map / filter / for-of が
// TypeError で落ちる（FRESTYLE-70 で staging 実機で観測された事象）。handler は
// repository の戻り値をそのまま c.JSON に渡す経路があるため、persistence 層で
// 空スライスを保証する必要がある。
func assertEmptyJSONArray[T any](t *testing.T, got []T, err error, name string) {
	t.Helper()
	require.NoError(t, err, "%s: エラーなく取得できること", name)
	assert.NotNil(t, got, "%s: 0 件でも nil スライスを返さないこと", name)
	assert.Empty(t, got, "%s: 0 件であること", name)

	encoded, marshalErr := json.Marshal(got)
	require.NoError(t, marshalErr, "%s: JSON 化できること", name)
	assert.Equal(t, "[]", string(encoded), "%s: JSON が null ではなく [] になること", name)
}

// TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration は、
// 一覧を返す repository メソッドが「該当行なし」で nil を返さないことを実 DB で検証する。
//
// 新規ユーザー / 新規コース / 未提出演習など、データがまだ無い状態は
// 新メンバーが最初に踏む動線であり、ここで画面が壊れると使い始められない（FRESTYLE-77）。
func TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("AI チャットのセッション一覧", func(t *testing.T) {
		got, err := persistence.NewAiChatSessionRepository(db).ListByUserID(ctx, noSuchID)
		assertEmptyJSONArray(t, got, err, "AiChatSessionRepository.ListByUserID")
	})

	t.Run("監査ログ一覧", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "audit_events")
		got, err := persistence.NewAuditRepository(db).ListRecent(ctx, 10)
		assertEmptyJSONArray(t, got, err, "AuditRepository.ListRecent")
	})

	t.Run("招待一覧（全体）", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "admin_invitations")
		got, err := persistence.NewAdminInvitationRepository(db).ListAll(ctx)
		assertEmptyJSONArray(t, got, err, "AdminInvitationRepository.ListAll")
	})

	t.Run("招待一覧（会社別）", func(t *testing.T) {
		got, err := persistence.NewAdminInvitationRepository(db).ListByCompanyID(ctx, noSuchID)
		assertEmptyJSONArray(t, got, err, "AdminInvitationRepository.ListByCompanyID")
	})

	t.Run("演習の提出履歴", func(t *testing.T) {
		got, err := persistence.NewExerciseSubmissionRepository(db).
			ListByUserAndExercise(ctx, noSuchID, noSuchID, domain.ExerciseKindMaster)
		assertEmptyJSONArray(t, got, err, "ExerciseSubmissionRepository.ListByUserAndExercise")
	})

	t.Run("章の進捗一覧", func(t *testing.T) {
		got, err := persistence.NewLessonProgressRepository(db).ListByUser(ctx, noSuchID)
		assertEmptyJSONArray(t, got, err, "LessonProgressRepository.ListByUser")
	})

	t.Run("教材一覧（会社別）", func(t *testing.T) {
		got, err := persistence.NewTeachingMaterialRepository(db).ListByCompany(ctx, noSuchID, true)
		assertEmptyJSONArray(t, got, err, "TeachingMaterialRepository.ListByCompany")
	})

	t.Run("教材一覧（コース別）", func(t *testing.T) {
		got, err := persistence.NewTeachingMaterialRepository(db).ListByCourse(ctx, noSuchID, true)
		assertEmptyJSONArray(t, got, err, "TeachingMaterialRepository.ListByCourse")
	})

	t.Run("演習一覧（言語別）", func(t *testing.T) {
		got, err := persistence.NewMasterExerciseRepository(db).
			ListByLanguage(ctx, "存在しない言語")
		assertEmptyJSONArray(t, got, err, "MasterExerciseRepository.ListByLanguage")
	})

	t.Run("演習の言語別集計", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "master_exercises")
		got, err := persistence.NewMasterExerciseRepository(db).SummaryByLanguage(ctx, noSuchID)
		assertEmptyJSONArray(t, got, err, "MasterExerciseRepository.SummaryByLanguage")
	})

	t.Run("学習レポート一覧", func(t *testing.T) {
		got, err := persistence.NewLearningReportRepository(db).ListByUserID(ctx, noSuchID)
		assertEmptyJSONArray(t, got, err, "LearningReportRepository.ListByUserID")
	})

	t.Run("日次の学習活動", func(t *testing.T) {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 3, 0)
		got, err := persistence.NewUserDailyActivityRepository(db).ListByUser(ctx, noSuchID, from, to)
		assertEmptyJSONArray(t, got, err, "UserDailyActivityRepository.ListByUser")
	})

	t.Run("最近見た章", func(t *testing.T) {
		got, err := persistence.NewUserChapterViewRepository(db).ListRecentByUser(ctx, noSuchID, 10)
		assertEmptyJSONArray(t, got, err, "UserChapterViewRepository.ListRecentByUser")
	})

	t.Run("会社ごとの在籍数", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users")
		got, err := persistence.NewCompanyStatsRepository(db).CountMembersByCompany(ctx)
		assertEmptyJSONArray(t, got, err, "CompanyStatsRepository.CountMembersByCompany")
	})
}
