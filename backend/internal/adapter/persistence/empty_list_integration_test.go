//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// noSuchID はどのテーブルにも該当行が無い状態を作るための存在しない ID。
const noSuchID uint64 = 999_999_999

// noSuchWorkspaceID は noSuchID の uuid 版（存在しないワークスペース）。
const noSuchWorkspaceID = "0198a000-0000-7000-8000-0000000000ff"

// listCase は「一覧を返す repository メソッド」1 件分の検証定義。
//
// call は any のスライスを返す形に揃える（要素型が異なるため）。
// truncate は 0 件の状態を作るために空にするテーブル（絞り込みで 0 件にできる場合は不要）。
type listCase struct {
	name     string
	truncate []string
	call     func(ctx context.Context, db *sql.DB) (any, error)
}

func listCases() []listCase {
	return []listCase{
		{
			name:     "監査ログ一覧",
			truncate: []string{"audit_events"},
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewAuditRepository(db).ListRecent(ctx, 10)
			},
		},
		{
			name: "招待一覧（全体）",
			// domain.AdminInvitation の TableName() は "invitations"（型名と一致しない）。
			truncate: []string{"invitations"},
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewAdminInvitationRepository(db).ListAll(ctx)
			},
		},
		{
			name: "招待一覧（会社別）",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewAdminInvitationRepository(db).ListByCompanyID(ctx, noSuchID)
			},
		},
		{
			name: "演習の提出履歴",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewExerciseSubmissionRepository(db).
					ListByUserAndExercise(ctx, noSuchID, noSuchID, domain.ExerciseKindMaster)
			},
		},
		{
			name: "章の進捗一覧",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewLessonProgressRepository(db).ListByUser(ctx, noSuchID)
			},
		},
		{
			name: "教材一覧（会社別）",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewTeachingMaterialRepository(db).ListByCompany(ctx, noSuchID, true)
			},
		},
		{
			name: "教材一覧（コース別）",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewTeachingMaterialRepository(db).ListByCourse(ctx, noSuchID, true)
			},
		},
		{
			name: "演習一覧（言語別）",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewMasterExerciseRepository(db).ListByLanguage(ctx, "存在しない言語")
			},
		},
		{
			name:     "演習の言語別集計",
			truncate: []string{"master_exercises"},
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewMasterExerciseRepository(db).SummaryByLanguage(ctx, noSuchID)
			},
		},
		{
			name: "日次の学習活動",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
				return persistence.NewUserDailyActivityRepository(db).
					ListByUser(ctx, noSuchID, from, from.AddDate(0, 3, 0))
			},
		},
		{
			name: "最近見た章",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewUserChapterViewRepository(db).ListRecentByUser(ctx, noSuchID, 10)
			},
		},
		{
			name:     "会社ごとの在籍数",
			truncate: []string{"users"},
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewCompanyStatsRepository(db).CountMembersByCompany(ctx)
			},
		},
		{
			name: "会社の在籍ユーザー一覧",
			call: func(ctx context.Context, db *sql.DB) (any, error) {
				return persistence.NewUserRepository(db).ListByWorkspaceID(ctx, noSuchWorkspaceID)
			},
		},
	}
}

// TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration は、
// 一覧を返す repository メソッドが「該当行なし」で nil を返さないことを実 DB で検証する。
//
// nil スライスは encoding/json で null になり、フロントの map / filter / for-of が
// TypeError で落ちる（FRESTYLE-70 で staging 実機で観測）。新規ユーザー・新規コース・
// 未提出演習という、新メンバーが最初に踏む動線で発生するため影響が大きい（FRESTYLE-77）。
func TestPersistence_一覧が0件でもnullではなく空配列を返すこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range listCases() {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.truncate) > 0 {
				testsupport.TruncateAll(t, sqlDB, tc.truncate...)
			}

			got, err := tc.call(ctx, sqlDB)

			require.NoError(t, err)
			assert.NotNil(t, got, "0 件でも nil スライスを返さないこと")
			encoded, marshalErr := json.Marshal(got)
			require.NoError(t, marshalErr)
			assert.Equal(t, "[]", string(encoded), "JSON が null ではなく [] になること")
		})
	}
}

// TestPersistence_一覧取得の失敗はエラーとして返ること_Integration は、
// 空配列の契約が「エラーを握り潰して空配列を返す」ことにならないよう異常系を固定する。
//
// 取得できなかったことを空配列として返すと、利用者には「0 件」と区別がつかず
// 障害に気づけなくなる。context を中断した状態で必ずエラーが返ることを確認する。
func TestPersistence_一覧取得の失敗はエラーとして返ること_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	canceled, cancel := context.WithCancel(context.Background())
	cancel() // 実行前に中断しておく

	for _, tc := range listCases() {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.call(canceled, sqlDB)

			assert.Error(t, err, "取得に失敗したらエラーを返すこと（空配列で握り潰さない）")
		})
	}
}
