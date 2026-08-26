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
)

// TestNoteRepository_Integration は NoteRepository の所有権スコープと並び順を実 Postgres で検証する。
func TestNoteRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewNoteRepository(sqlDB)
	ctx := context.Background()

	t.Run("ListByUserID は自分の note だけを updated_at DESC で返す", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notes")

		older := &domain.Note{UserID: 7, Title: "older", UpdatedAt: time.Now().Add(-time.Hour)}
		newer := &domain.Note{UserID: 7, Title: "newer", UpdatedAt: time.Now()}
		other := &domain.Note{UserID: 8, Title: "someone-else"}
		require.NoError(t, repo.Create(ctx, older))
		require.NoError(t, repo.Create(ctx, newer))
		require.NoError(t, repo.Create(ctx, other))

		rows, err := repo.ListByUserID(ctx, 7)
		require.NoError(t, err)
		require.Len(t, rows, 2, "user 8 の note は WHERE user_id で除外される")
		require.Equal(t, "newer", rows[0].Title, "updated_at DESC で新しい順")
		require.Equal(t, "older", rows[1].Title)
	})

	t.Run("Delete は user_id スコープで他人の note を消さない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notes")

		mine := &domain.Note{UserID: 7, Title: "mine"}
		theirs := &domain.Note{UserID: 8, Title: "theirs"}
		require.NoError(t, repo.Create(ctx, mine))
		require.NoError(t, repo.Create(ctx, theirs))

		// user 7 が user 8 の note を消そうとしても WHERE user_id=7 で 0 行になる。
		// 期待値を nil から domain.ErrNotFound へ更新した理由:
		//   0 行削除まで成功にしていると、呼び出し側は「1 件消した」と「何も起きなかった」を
		//   区別できず、削除が効いていないのに画面から行が消える。
		//   「他人の note」と「存在しない id」はどちらも 0 行 = 同じ domain.ErrNotFound に畳まれるので、
		//   存在オラクル（FRESTYLE-367 / 376 で塞いだ穴）は開かない。
		require.ErrorIs(t, repo.Delete(ctx, 7, theirs.ID), domain.ErrNotFound)
		got, err := repo.FindByID(ctx, 8, theirs.ID)
		require.NoError(t, err, "他人の note は残る")
		require.Equal(t, "theirs", got.Title)

		// 自分の note は消せる。
		require.NoError(t, repo.Delete(ctx, 7, mine.ID))
		_, err = repo.FindByID(ctx, 7, mine.ID)
		require.Error(t, err, "自分の note は削除済み")
	})

	// SQL 側の多層防御を直接固定する。usecase の所有者判定を通さず repository を叩き、
	// 「他人の note」と「存在しない id」が同じ domain.ErrNotFound になることを見る。
	// WHERE の user_id 述語を外すと、他人の note が 1 行返ってこのテストが落ちる。
	t.Run("FindByID は他人の note を存在しない id と同じ扱いにする", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notes")

		theirs := &domain.Note{UserID: 8, Title: "theirs", Content: "secret"}
		require.NoError(t, repo.Create(ctx, theirs))

		_, foreignErr := repo.FindByID(ctx, 7, theirs.ID)
		require.ErrorIs(t, foreignErr, domain.ErrNotFound, "他人の note は SQL の時点で 0 行")

		// 実在しない id（採番済み最大値より先）を引いたときと同じエラーであること。
		_, missingErr := repo.FindByID(ctx, 7, theirs.ID+1_000_000)
		require.ErrorIs(t, missingErr, domain.ErrNotFound)
		require.Equal(t, missingErr.Error(), foreignErr.Error(), "他人と不在でエラーが撃ち分けられている")

		// 所有者本人からは当然読める（述語が効きすぎていないことの裏取り）。
		got, err := repo.FindByID(ctx, 8, theirs.ID)
		require.NoError(t, err)
		require.Equal(t, "theirs", got.Title)
	})

	t.Run("Update は内容を保存する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notes")

		n := &domain.Note{UserID: 7, Title: "before", Content: "x"}
		require.NoError(t, repo.Create(ctx, n))
		n.Title = "after"
		require.NoError(t, repo.Update(ctx, n))

		got, err := repo.FindByID(ctx, 7, n.ID)
		require.NoError(t, err)
		require.Equal(t, "after", got.Title)
	})

	// 書き込み側の多層防御。UPDATE 文の user_id 述語を外すと、他人の note が
	// 上書きされてこのテストが落ちる。
	t.Run("Update は他人の note を書き換えない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "notes")

		theirs := &domain.Note{UserID: 8, Title: "theirs", Content: "original"}
		require.NoError(t, repo.Create(ctx, theirs))

		// 所有者を偽って（user 7 として）他人の note を更新しようとする。
		forged := &domain.Note{ID: theirs.ID, UserID: 7, Title: "hijacked", Content: "overwritten"}
		require.ErrorIs(t, repo.Update(ctx, forged), domain.ErrNotFound, "0 行更新は not found")

		got, err := repo.FindByID(ctx, 8, theirs.ID)
		require.NoError(t, err)
		require.Equal(t, "theirs", got.Title, "他人の note は書き換わらない")
		require.Equal(t, "original", got.Content)
	})
}

// TestNoteRepository_Timestamps_Integration は Create が created_at / updated_at を now で埋め、
// Update が updated_at を進めつつ created_at を保つこと（GORM autoTime 相当）を固定する。
func TestNoteRepository_Timestamps_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewNoteRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "notes")

	before := time.Now().Add(-time.Second)
	n := &domain.Note{UserID: 7, Title: "before", Content: "x"}
	require.NoError(t, repo.Create(ctx, n))
	require.False(t, n.CreatedAt.IsZero(), "created_at が書き戻される")
	require.False(t, n.UpdatedAt.IsZero(), "updated_at が書き戻される")
	require.True(t, n.CreatedAt.After(before), "created_at は now 付近")
	createdAt := n.CreatedAt

	time.Sleep(5 * time.Millisecond)
	n.Title = "after"
	require.NoError(t, repo.Update(ctx, n))

	got, err := repo.FindByID(ctx, 7, n.ID)
	require.NoError(t, err)
	require.Equal(t, "after", got.Title)
	require.WithinDuration(t, createdAt, got.CreatedAt, time.Millisecond, "created_at は更新で変わらない")
	require.True(t, got.UpdatedAt.After(createdAt), "updated_at は更新で進む")
}
