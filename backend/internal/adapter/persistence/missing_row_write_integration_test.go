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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// missingID は「int64 に収まるが、どのテーブルにも存在しない」id。
// 採番列（bigserial）は 1 から増えるので、TRUNCATE ... RESTART IDENTITY 済みの DB では
// ここまで飛ばせば確実に空き番になる。
const missingID uint64 = 9_000_000_000

// missingRowTables は下の表で触るテーブル。TRUNCATE で毎回まっさらにしてから叩く
// （既存行に偶然当たって「0 行ではなかった」ことを見落とさないため）。
var missingRowTables = []string{
	"notifications",
	"users",
}

// missingRowCase は「存在しない行を狙う書き込み」1 件分の検証定義。
type missingRowCase struct {
	name string
	call func(ctx context.Context, db *sql.DB) error
}

// missingRowWriteCases は Create 以外の書き込み（Update / Delete / MarkRead 等）のうち、
// 対象行が 0 行だったときに not-found を返すべきものを列挙する。
//
// なぜ 0 行を成功にしてはいけないか:
//
//	UPDATE / DELETE は 1 行も一致しなくても SQL としては成功する（GORM 時代から
//	`.Error == nil` になり、sqlc でも :exec は「エラーなく流れたか」しか返さない）。
//	repository が nil を返すと usecase は「保存できた」と判断し、handler は 200 / 204 を返す。
//	利用者の画面には保存済みと表示されるのに DB には何も書かれていない、という
//	取り違えがそのまま外へ出る。ここを domain.ErrNotFound に畳んで 404 に揃える。
func missingRowWriteCases() []missingRowCase {
	return []missingRowCase{
		{
			name: "通知の既読化",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewNotificationRepository(db).MarkRead(ctx, 7, missingID)
			},
		},
		{
			name: "user の氏名更新",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserRepository(db).UpdateName(ctx, missingID, "新しい名前")
			},
		},
		{
			name: "user の所属ワークスペース付け替え",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserRepository(db).UpdateWorkspaceID(ctx, missingID, nil)
			},
		},
		{
			name: "user の有効/無効更新",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserRepository(db).UpdateActive(ctx, missingID, false)
			},
		},
		{
			name: "user の論理削除",
			call: func(ctx context.Context, db *sql.DB) error {
				return persistence.NewUserRepository(db).SoftDelete(ctx, missingID)
			},
		},
	}
}

// TestPersistence_存在しない行への書き込みはnot_foundを返すこと_Integration は
// Create 以外の書き込みが「0 行でも成功」を返さないことを実 PostgreSQL 上で固定する。
//
// 0 行を成功にしていた頃は、handler が 200 / 204 を返し、呼び出し側は更新できたと誤認していた
// （利用者には保存済みと見えて、実際は 1 行も書かれていない）。ここが緑である限り、
// どのメソッドも「対象が無い」を 404 として上位へ伝える。
func TestPersistence_存在しない行への書き込みはnot_foundを返すこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, tc := range missingRowWriteCases() {
		t.Run(tc.name, func(t *testing.T) {
			testsupport.TruncateAll(t, sqlDB, missingRowTables...)

			err := tc.call(ctx, sqlDB)

			require.Error(t, err, "0 行の書き込みを成功として返している（呼び出し側が保存できたと誤認する）")
			assert.ErrorIs(t, err, domain.ErrNotFound,
				"not-found は domain.ErrNotFound で返す（handler が 404 に分岐する契約）")
		})
	}
}

// TestPersistence_一括操作は0件でも成功のままであること_Integration は、
// 単一行を狙う書き込みと違って「対象を全部畳む」一括操作は 0 件でも成功のままであることを固定する。
//
// ここを not-found にすると、未読が 1 件も無い状態で「すべて既読にする」を押しただけ、
// 章が 1 つも無いコースを消しただけ、まだ完了していない章のチェックを外しただけで 404 が返る。
// 「対象が無い」ことが異常ではない操作まで巻き込まないための線引きをテストで残す。
func TestPersistence_一括操作は0件でも成功のままであること_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, missingRowTables...)

	t.Run("未読が無い user の一括既読化は成功", func(t *testing.T) {
		require.NoError(t, persistence.NewNotificationRepository(sqlDB).MarkAllRead(ctx, missingID))
	})
}

// TestKnowledgeBase_存在しないページのアーカイブ操作はnot_foundを返すこと_Integration は
// ノート側（UUID 主キー・:execrows の戻り値を捨てていた経路）を同じ観点で固定する。
//
// ノートは domain.ErrNotFound ではなく repository.ErrPageNotFound を使う
// （handler の respondKnowledgeBaseErr が「存在しない」と「権限が無い」を同じ 404 に畳むため）。
func TestKnowledgeBase_存在しないページのアーカイブ操作はnot_foundを返すこと_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewKnowledgeBaseRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, kbTables...)

	ws := createWorkspace(t, sqlDB, "ws-missing-row")
	space := createSpace(t, sqlDB, ws, "mrw")
	// 実在するページを 1 枚だけ作る。存在する経路が壊れていないことも同時に見る。
	uc := newKbUseCases(repo)
	live := mustCreatePage(ctx, t, uc, ws, space, nil, "生きているページ")

	// 形式は正しいが、このワークスペースに存在しないページ ID。
	const absentPageID = "00000000-0000-4000-8000-0000000000ff"

	t.Run("存在しないページのアーカイブは not-found", func(t *testing.T) {
		err := repo.ArchivePageSubtree(ctx, ws, absentPageID)
		require.Error(t, err, "0 行更新を成功として返している（ツリーから消えたと誤認する）")
		assert.ErrorIs(t, err, repository.ErrPageNotFound)
	})

	t.Run("存在するページのアーカイブは成功する", func(t *testing.T) {
		require.NoError(t, repo.ArchivePageSubtree(ctx, ws, live.ID))
		archived, err := repo.FindPage(ctx, ws, live.ID)
		require.NoError(t, err)
		require.NotNil(t, archived.ArchivedAt, "実際にアーカイブされている")

		// 復帰も同じ経路で確かめる（0 行なら not-found、1 行なら成功）。
		require.NoError(t, repo.UnarchivePageSubtree(ctx, ws, live.ID, *archived.ArchivedAt, nil))
		restored, err := repo.FindPage(ctx, ws, live.ID)
		require.NoError(t, err)
		require.Nil(t, restored.ArchivedAt, "実際に現役へ戻っている")
	})

	t.Run("存在しないページの復帰は not-found", func(t *testing.T) {
		err := repo.UnarchivePageSubtree(ctx, ws, absentPageID, live.CreatedAt, nil)
		require.Error(t, err, "0 行更新を成功として返している（復帰したと誤認する）")
		assert.ErrorIs(t, err, repository.ErrPageNotFound)
	})
}
