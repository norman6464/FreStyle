//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestSessionNoteRepository_CrossUserOverwrite_Integration は、他人のセッションメモを
// 上書きできないことを実 PostgreSQL で固定する。
//
// session_id に一意制約を張って真の upsert にした際、ON CONFLICT DO UPDATE に所有者の
// 条件が無かったため、他人が同じ session_id へ書くと既存行（被害者のもの）の content だけが
// 置き換わり、user_id は被害者のまま残っていた。被害者からは自分のメモが書き換わったように
// しか見えず、書いた人は記録に残らない。
func TestSessionNoteRepository_CrossUserOverwrite_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(db)
	ctx := context.Background()

	const (
		sessionID = uint64(9001)
		owner     = uint64(8)
		attacker  = uint64(7)
	)

	setup := func(t *testing.T) {
		t.Helper()
		testsupport.TruncateAll(t, db, "session_notes")
		require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{
			SessionID: sessionID, UserID: owner, Content: "所有者のメモ",
		}))
	}

	t.Run("他人は既存メモを上書きできず not-found になる", func(t *testing.T) {
		setup(t)

		err := repo.Upsert(ctx, &domain.SessionNote{
			SessionID: sessionID, UserID: attacker, Content: "攻撃者が上書き",
		})
		require.ErrorIs(t, err, domain.ErrNotFound)

		got, err := repo.FindBySessionID(ctx, sessionID)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "所有者のメモ", got.Content, "他人の書き込みで内容が変わってはいけない")
		require.Equal(t, owner, got.UserID, "所有者が移ってはいけない")
	})

	t.Run("所有者は自分のメモを更新できる", func(t *testing.T) {
		setup(t)

		require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{
			SessionID: sessionID, UserID: owner, Content: "本人が更新",
		}))

		got, err := repo.FindBySessionID(ctx, sessionID)
		require.NoError(t, err)
		require.Equal(t, "本人が更新", got.Content)
		require.Equal(t, owner, got.UserID)
	})

	t.Run("他人が書いても行は増えない", func(t *testing.T) {
		setup(t)

		// 戻り値を捨てない。拒否されたこと（not-found）まで確かめないと、実装が変わって
		// 書き込みが通るようになっても「行数 1」だけは満たされ、このテストは緑のまま通る
		// （＝上書きに戻った回帰を見逃す）。
		err := repo.Upsert(ctx, &domain.SessionNote{
			SessionID: sessionID, UserID: attacker, Content: "攻撃者が上書き",
		})
		require.ErrorIs(t, err, domain.ErrNotFound)

		sqlDB := db
		var n int
		require.NoError(t, sqlDB.QueryRowContext(ctx,
			`SELECT count(*) FROM session_notes WHERE session_id = $1`, sessionID).Scan(&n))
		require.Equal(t, 1, n)
	})
}
