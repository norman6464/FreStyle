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
	"gorm.io/gorm"
)

// aiChatSessionRow は ai_chat_sessions の 1 行を repository を通さず直に読むための型。
// 「repository が実際にどの列へ書いたか」を repository 自身の読み取りに頼らず確かめる。
type aiChatSessionRow struct {
	ID          uint64
	UserID      uint64
	Title       string
	SessionType string
	ScenarioID  *uint64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func readAiChatSessionRow(t *testing.T, db *gorm.DB, id uint64) aiChatSessionRow {
	t.Helper()
	var row aiChatSessionRow
	require.NoError(t, db.Raw(
		"SELECT id, user_id, title, session_type, scenario_id, created_at, updated_at FROM ai_chat_sessions WHERE id = ?",
		id,
	).Scan(&row).Error)
	return row
}

// TestAiChatSessionRepository_Create_Integration は Create の契約
// （採番 id と時刻の書き戻し / NULL 可の scenario_id / 明示した時刻は保持）を固定する。
func TestAiChatSessionRepository_Create_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewAiChatSessionRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "ai_chat_sessions")

	t.Run("採番 id と created_at / updated_at を書き戻す", func(t *testing.T) {
		scenario := uint64(42)
		s := &domain.AiChatSession{
			UserID:      7,
			Title:       "はじめてのチャット",
			SessionType: domain.AiChatSessionTypePractice,
			ScenarioID:  &scenario,
		}
		before := time.Now()
		require.NoError(t, repo.Create(ctx, s))
		require.NotZero(t, s.ID, "採番された id が書き戻る")
		require.False(t, s.CreatedAt.IsZero(), "created_at が入る（DB 既定値は無い）")
		require.False(t, s.UpdatedAt.IsZero(), "updated_at が入る（DB 既定値は無い）")
		require.False(t, s.CreatedAt.Before(before.Add(-time.Minute)), "created_at は現在時刻")

		row := readAiChatSessionRow(t, db, s.ID)
		require.Equal(t, uint64(7), row.UserID)
		require.Equal(t, "はじめてのチャット", row.Title)
		require.Equal(t, domain.AiChatSessionTypePractice, row.SessionType)
		require.NotNil(t, row.ScenarioID)
		require.Equal(t, uint64(42), *row.ScenarioID)
		require.WithinDuration(t, s.CreatedAt, row.CreatedAt, time.Second)
		require.WithinDuration(t, s.UpdatedAt, row.UpdatedAt, time.Second)
	})

	t.Run("scenario_id 未指定は NULL で入る", func(t *testing.T) {
		s := &domain.AiChatSession{UserID: 8, Title: "free", SessionType: domain.AiChatSessionTypeFree}
		require.NoError(t, repo.Create(ctx, s))
		row := readAiChatSessionRow(t, db, s.ID)
		require.Nil(t, row.ScenarioID)
	})

	t.Run("明示した created_at / updated_at は上書きされない", func(t *testing.T) {
		fixed := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
		s := &domain.AiChatSession{
			UserID: 9, Title: "fixed", SessionType: domain.AiChatSessionTypeFree,
			CreatedAt: fixed, UpdatedAt: fixed,
		}
		require.NoError(t, repo.Create(ctx, s))
		row := readAiChatSessionRow(t, db, s.ID)
		require.WithinDuration(t, fixed, row.CreatedAt, time.Second)
		require.WithinDuration(t, fixed, row.UpdatedAt, time.Second)
	})
}

// TestAiChatSessionRepository_Read_Integration は ListByUserID / FindByID の契約
// （user_id の絞り / created_at DESC, id DESC のタイブレーク / 未存在の not found）を固定する。
func TestAiChatSessionRepository_Read_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewAiChatSessionRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "ai_chat_sessions")

	// 同一 created_at の 3 件（タイブレーク検証用）+ より新しい 1 件 + 他人の 1 件。
	same := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newer := same.Add(time.Hour)
	tied := make([]uint64, 0, 3)
	for i := 0; i < 3; i++ {
		s := &domain.AiChatSession{UserID: 7, Title: "tied", SessionType: domain.AiChatSessionTypeFree, CreatedAt: same, UpdatedAt: same}
		require.NoError(t, repo.Create(ctx, s))
		tied = append(tied, s.ID)
	}
	newest := &domain.AiChatSession{UserID: 7, Title: "newest", SessionType: domain.AiChatSessionTypeFree, CreatedAt: newer, UpdatedAt: newer}
	require.NoError(t, repo.Create(ctx, newest))
	other := &domain.AiChatSession{UserID: 8, Title: "other", SessionType: domain.AiChatSessionTypeFree, CreatedAt: newer, UpdatedAt: newer}
	require.NoError(t, repo.Create(ctx, other))

	t.Run("自分のセッションだけを created_at DESC, id DESC で返す", func(t *testing.T) {
		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		got := make([]uint64, 0, len(rows))
		for _, r := range rows {
			require.Equal(t, uint64(7), r.UserID, "他人のセッションは含まない")
			got = append(got, r.ID)
		}
		// newest が先頭。同時刻の 3 件は id 降順で決定的に並ぶ。
		require.Equal(t, []uint64{newest.ID, tied[2], tied[1], tied[0]}, got)
	})

	t.Run("該当なしは空スライス（nil ではない）", func(t *testing.T) {
		rows, err := repo.ListByUserID(ctx, noSuchID)
		require.NoError(t, err)
		require.NotNil(t, rows)
		require.Empty(t, rows)
	})

	t.Run("FindByID は全列を詰めて返す", func(t *testing.T) {
		got, err := repo.FindByID(ctx, newest.ID)
		require.NoError(t, err)
		require.Equal(t, newest.ID, got.ID)
		require.Equal(t, uint64(7), got.UserID)
		require.Equal(t, "newest", got.Title)
		require.Equal(t, domain.AiChatSessionTypeFree, got.SessionType)
		require.Nil(t, got.ScenarioID)
		require.WithinDuration(t, newer, got.CreatedAt, time.Second)
		require.WithinDuration(t, newer, got.UpdatedAt, time.Second)
	})

	t.Run("FindByID の未存在は domain.ErrNotFound", func(t *testing.T) {
		got, err := repo.FindByID(ctx, noSuchID)
		require.ErrorIs(t, err, domain.ErrNotFound, "handler が 404 に分岐するシグナル")
		require.Nil(t, got)
	})
}

// TestAiChatSessionRepository_UpdateTitle_Integration は UpdateTitle の契約
// （書くのは title と updated_at だけ / 他行は触らない / 存在しない id はエラーにしない）を固定する。
func TestAiChatSessionRepository_UpdateTitle_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewAiChatSessionRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "ai_chat_sessions")

	scenario := uint64(99)
	old := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	target := &domain.AiChatSession{
		UserID: 7, Title: "before", SessionType: domain.AiChatSessionTypePractice,
		ScenarioID: &scenario, CreatedAt: old, UpdatedAt: old,
	}
	require.NoError(t, repo.Create(ctx, target))
	bystander := &domain.AiChatSession{UserID: 7, Title: "bystander", SessionType: domain.AiChatSessionTypeFree, CreatedAt: old, UpdatedAt: old}
	require.NoError(t, repo.Create(ctx, bystander))

	require.NoError(t, repo.UpdateTitle(ctx, target.ID, "after"))

	t.Run("title を書き updated_at を進める", func(t *testing.T) {
		row := readAiChatSessionRow(t, db, target.ID)
		require.Equal(t, "after", row.Title)
		require.True(t, row.UpdatedAt.After(old), "updated_at は現在時刻へ進む")
		require.WithinDuration(t, time.Now(), row.UpdatedAt, time.Minute)
	})

	t.Run("title / updated_at 以外の列は変えない", func(t *testing.T) {
		row := readAiChatSessionRow(t, db, target.ID)
		require.Equal(t, uint64(7), row.UserID)
		require.Equal(t, domain.AiChatSessionTypePractice, row.SessionType)
		require.NotNil(t, row.ScenarioID)
		require.Equal(t, uint64(99), *row.ScenarioID)
		require.WithinDuration(t, old, row.CreatedAt, time.Second, "created_at は不変")
	})

	t.Run("対象外の行は触らない", func(t *testing.T) {
		row := readAiChatSessionRow(t, db, bystander.ID)
		require.Equal(t, "bystander", row.Title)
		require.WithinDuration(t, old, row.UpdatedAt, time.Second)
	})

	t.Run("存在しない id への UpdateTitle はエラーにしない（0 行更新）", func(t *testing.T) {
		require.NoError(t, repo.UpdateTitle(ctx, noSuchID, "ghost"))
	})
}

// TestAiChatSessionRepository_Delete_Integration は Delete の契約
// （物理削除であること / 対象外の行を消さない / 存在しない id はエラーにしない）を固定する。
func TestAiChatSessionRepository_Delete_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewAiChatSessionRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "ai_chat_sessions")

	t.Run("ai_chat_sessions は soft delete 列を持たない", func(t *testing.T) {
		var cols []string
		require.NoError(t, db.Raw(
			"SELECT column_name FROM information_schema.columns WHERE table_name = 'ai_chat_sessions'",
		).Scan(&cols).Error)
		require.NotContains(t, cols, "deleted_at", "deleted_at が無い = Delete は物理削除")
	})

	target := &domain.AiChatSession{UserID: 7, Title: "doomed", SessionType: domain.AiChatSessionTypeFree}
	require.NoError(t, repo.Create(ctx, target))
	survivor := &domain.AiChatSession{UserID: 7, Title: "survivor", SessionType: domain.AiChatSessionTypeFree}
	require.NoError(t, repo.Create(ctx, survivor))

	require.NoError(t, repo.Delete(ctx, target.ID))

	t.Run("行そのものが消える（論理削除ではない）", func(t *testing.T) {
		var cnt int64
		require.NoError(t, db.Raw("SELECT count(*) FROM ai_chat_sessions WHERE id = ?", target.ID).Scan(&cnt).Error)
		require.Equal(t, int64(0), cnt, "物理削除なので行が残らない")

		_, err := repo.FindByID(ctx, target.ID)
		require.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("対象外の行は残る", func(t *testing.T) {
		got, err := repo.FindByID(ctx, survivor.ID)
		require.NoError(t, err)
		require.Equal(t, "survivor", got.Title)
	})

	t.Run("存在しない id への Delete はエラーにしない（0 行削除）", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, noSuchID))
	})
}
