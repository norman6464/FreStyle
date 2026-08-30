//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
)

// outOfRangeID は bigint(int64) に収まらない uint64。DB の採番列からは出てこないが、
// 書き込み系がこの値を渡されたときに「何も書かずに成功」を返さないことを固定するために使う。
const outOfRangeID uint64 = math.MaxUint64

// writeCase は「行を書き込む repository メソッド」1 件分の検証定義。
type writeCase struct {
	name string
	call func(ctx context.Context, db *sql.DB) error
}

func outOfRangeWriteCases() []writeCase {
	return []writeCase{
		{
			name: "コースの作成（created_by_user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewCourseRepository(db).Create(ctx, &domain.Course{
					CreatedByUserID: outOfRangeID, Title: "t",
				})
			},
		},
		{
			name: "演習提出の作成（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewExerciseSubmissionRepository(db).Create(ctx, &domain.ExerciseSubmission{
					UserID: outOfRangeID, ExerciseKind: domain.ExerciseKindMaster, ExerciseID: 1,
					SubmittedAt: time.Now(),
				})
			},
		},
		{
			name: "演習提出の作成（exercise_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewExerciseSubmissionRepository(db).Create(ctx, &domain.ExerciseSubmission{
					UserID: 1, ExerciseKind: domain.ExerciseKindMaster, ExerciseID: outOfRangeID,
					SubmittedAt: time.Now(),
				})
			},
		},
		{
			name: "ノートの作成（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewNoteRepository(db).Create(ctx, &domain.Note{
					UserID: outOfRangeID, Title: "t", Content: "c",
				})
			},
		},
		{
			name: "通知の作成（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewNotificationRepository(db).Create(ctx, &domain.Notification{
					UserID: outOfRangeID, Type: "info", Title: "t", Body: "b",
				})
			},
		},
		{
			name: "プロフィールの upsert（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewProfileRepository(db).Upsert(ctx, &domain.Profile{
					UserID: outOfRangeID, Bio: "b",
				})
			},
		},
		{
			name: "リッチ文書の作成（owner_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewRichDocumentRepository(db).Create(ctx, &domain.RichDocument{
					OwnerID: outOfRangeID, Kind: domain.DocumentKindNote, Title: "t",
					Doc: `{"type":"doc","content":[]}`,
				})
			},
		},
		{
			name: "章の作成（course_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewTeachingMaterialRepository(db).Create(ctx, &domain.TeachingMaterial{
					CourseID: outOfRangeID, CreatedByUserID: 1, Title: "t",
				})
			},
		},
		{
			name: "章の作成（created_by_user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewTeachingMaterialRepository(db).Create(ctx, &domain.TeachingMaterial{
					CourseID: 1, CreatedByUserID: outOfRangeID, Title: "t",
				})
			},
		},
		{
			name: "章の閲覧記録の upsert（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserChapterViewRepository(db).UpsertView(ctx, outOfRangeID, 1, 1)
			},
		},
		{
			name: "日次活動の加算（user_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserDailyActivityRepository(db).Increment(
					ctx, outOfRangeID, time.Now(), repository.UserDailyActivityIncrement{LessonCount: 1},
				)
			},
		},
	}
}

// TestPersistence_書き込みは範囲外idを成功として返さないこと_Integration は、
// bigint に収まらない id を渡された書き込み系が nil（成功）を返さないことを実 DB で固定する。
//
// 1 行も書いていないのに nil を返すと、呼び出し側は作成・更新できたと誤認する
// （usecase はそのまま 201 / 200 を返し、次の取得で「無い」ことに初めて気づく）。
// 読み取り系が「存在し得ない id = 0 件 / not found」を返すのとは扱いが異なる。
func TestPersistence_書き込みは範囲外idを成功として返さないこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range outOfRangeWriteCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(ctx, sqlDB)

			assert.Error(t, err, "書き込めていないのに成功を返さないこと")
		})
	}
}

// malformedWorkspaceID は UUID として解釈できない workspace_id。所属参照は bigint から
// uuid へ移ったため「範囲外の id」は存在しないが、値が壊れていて 1 行も書けない状況は残る。
const malformedWorkspaceID = "not-a-uuid"

// malformedWorkspaceWriteCases は所属参照（workspace_id）が壊れた値で渡される書き込み。
func malformedWorkspaceWriteCases() []writeCase {
	bad := malformedWorkspaceID
	return []writeCase{
		{
			name: "招待の作成（workspace_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewAdminInvitationRepository(db).Create(ctx, &domain.AdminInvitation{
					WorkspaceID: &bad, Email: "a@example.com", Role: domain.RoleCompanyAdmin,
					Status: domain.InvitationStatusPending, ExpiresAt: time.Now().Add(time.Hour),
				})
			},
		},
		{
			name: "コースの作成（workspace_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewCourseRepository(db).Create(ctx, &domain.Course{
					WorkspaceID: &bad, CreatedByUserID: 1, Title: "t",
				})
			},
		},
		{
			name: "章の作成（workspace_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewTeachingMaterialRepository(db).Create(ctx, &domain.TeachingMaterial{
					WorkspaceID: &bad, CourseID: 1, CreatedByUserID: 1, Title: "t",
				})
			},
		},
		{
			name: "リッチ文書の作成（workspace_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewRichDocumentRepository(db).Create(ctx, &domain.RichDocument{
					OwnerID: 1, WorkspaceID: &bad, Kind: domain.DocumentKindNote, Title: "t",
					Doc: `{"type":"doc","content":[]}`,
				})
			},
		},
		{
			name: "ユーザーの所属付け替え（workspace_id）",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserRepository(db).UpdateWorkspaceID(ctx, 1, &bad)
			},
		},
	}
}

// TestPersistence_書き込みは不正な形式のworkspace_idを成功として返さないこと_Integration は、
// UUID として読めない workspace_id を渡された書き込み系が nil（成功）を返さないことを固定する。
//
// ここを黙って NULL 扱いにすると、所属の付いていない行（誰からも見えない、あるいは
// 誰からも見える行）が「作成できた」という応答とともに残る。
func TestPersistence_書き込みは不正な形式のworkspace_idを成功として返さないこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range malformedWorkspaceWriteCases() {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(ctx, sqlDB)

			assert.Error(t, err, "書き込めていないのに成功を返さないこと")
		})
	}
}
