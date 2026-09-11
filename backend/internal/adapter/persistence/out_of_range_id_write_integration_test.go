//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"math"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// outOfRangeID は bigint(int64) に収まらない uint64。DB の採番列からは出てこないが、
// 書き込み系がこの値を渡されたときに「何も書かずに成功」を返さないことを固定するために使う。
const outOfRangeID uint64 = math.MaxUint64

// writeCase は「行を書き込む repository メソッド」1 件分の検証定義。
//
// column は返るエラーに含まれるはずの列名。ここまで具体的に固定するのは、
// workspace_id / page_id / thread_id を実在しない値のままにすると、範囲外 id の
// チェックより先に「参照先が無い」という別の理由で落ちてしまい、範囲外 id の
// チェック自体が無くなっても気づけない（テストが別の失敗で空振りする）ため。
type writeCase struct {
	name   string
	column string
	call   func(ctx context.Context, db *sql.DB) error
}

// outOfRangeWriteCases を組み立てる前に、実在するワークスペース・スペース・ページ・
// コメントスレッドを用意する。範囲外にするのは各ケースが検証する 1 列だけにする。
func outOfRangeWriteCases(t *testing.T, db *sql.DB) []writeCase {
	t.Helper()
	ws := createWorkspace(t, db, "ws-out-of-range-write")
	space := createSpace(t, db, ws, "eng")
	page := createPage(t, db, ws, space, nil, "a0")

	// スレッドの作成者は範囲外 id チェックの対象ではないので、有効な値（任意の bigint）にする。
	thread, err := persistence.NewCommentRepository(db).CreateCommentThread(
		context.Background(), ws, page, 1, repository.CommentAnchor{},
	)
	require.NoError(t, err)

	return []writeCase{
		{
			name:   "コメントスレッドの作成（created_by_user_id）",
			column: "created_by_user_id",
			call: func(ctx context.Context, db *sql.DB) error {
				_, err := persistence.NewCommentRepository(db).CreateCommentThread(
					ctx, ws, page, outOfRangeID, repository.CommentAnchor{},
				)
				return err
			},
		},
		{
			name:   "コメントの作成（author_user_id）",
			column: "author_user_id",
			call: func(ctx context.Context, db *sql.DB) error {
				_, err := persistence.NewCommentRepository(db).CreateComment(
					ctx, thread.ID, outOfRangeID, `[{"type":"text","text":"x"}]`,
				)
				return err
			},
		},
		{
			name:   "通知の作成（user_id）",
			column: "user_id",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewNotificationRepository(db).Create(ctx, &domain.Notification{
					UserID: outOfRangeID, Type: "info", Title: "t", Body: "b",
				})
			},
		},
		{
			name:   "プロフィールの upsert（user_id）",
			column: "user_id",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewProfileRepository(db).Upsert(ctx, &domain.Profile{
					UserID: outOfRangeID, Bio: "b",
				})
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
//
// エラーの中身が「範囲外 id」であることまで確かめる（tc.column を含むこと）。
// 参照先（workspace/page/thread）は実在するものを使い、範囲外 id チェックの前段にある
// 「参照先が無い」という別の失敗で空振りしないようにしている。
func TestPersistence_書き込みは範囲外idを成功として返さないこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range outOfRangeWriteCases(t, sqlDB) {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.call(ctx, sqlDB)

			require.Error(t, err, "書き込めていないのに成功を返さないこと")
			assert.ErrorContains(t, err, "bigint の範囲を超えている")
			assert.ErrorContains(t, err, tc.column, "エラーが対象の列を名指ししていること")
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
