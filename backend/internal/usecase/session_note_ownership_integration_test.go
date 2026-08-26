//go:build integration

package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/require"
)

// TestUpsertSessionNote_SessionOwnership_Integration は「セッションメモを書いてよいのは
// セッションの所有者だけ」という不変条件を実 PostgreSQL で固定する。
//
// メモの所有権はメモの行ではなくセッションに従属する。権威は ai_chat_sessions.user_id で、
// session_notes.user_id はその結果でしかない。
//
// SQL 側の防壁（ON CONFLICT (session_id) DO UPDATE ... WHERE session_notes.user_id =
// EXCLUDED.user_id）は衝突して UPDATE に進んだときにしか評価されないので、**メモがまだ無い
// セッションへの初回 INSERT は素通りする**。実際、この検証を入れる前は他人のセッションに
// 攻撃者名義のメモを新規作成でき、その結果:
//
//	被害者が読む → 行の所有者が違うので nil（メモが無いように見える）
//	被害者が書く → 既存行と衝突し WHERE で弾かれて 0 行 → not-found
//	           → 被害者は自分のセッションでメモを取れなくなる
//
// という「上書きより悪い」状態になっていた。ここではその新規作成の経路を主に固定する。
func TestUpsertSessionNote_SessionOwnership_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	notes := persistence.NewSessionNoteRepository(db)
	sessions := persistence.NewAiChatSessionRepository(db)
	upsert := usecase.NewUpsertSessionNoteUseCase(notes, sessions)
	get := usecase.NewGetSessionNoteUseCase(notes)

	const (
		owner    = uint64(8)
		attacker = uint64(7)
	)

	// newSession は owner が所有するセッションを 1 件作り、その id を返す。
	// 毎回テーブルを空にしてから作るので、サブテスト同士は互いの行を見ない。
	newSession := func(t *testing.T) uint64 {
		t.Helper()
		testsupport.TruncateAll(t, db, "session_notes", "ai_chat_sessions")
		s := &domain.AiChatSession{
			UserID:      owner,
			Title:       "被害者のセッション",
			SessionType: domain.AiChatSessionTypeFree,
		}
		require.NoError(t, sessions.Create(ctx, s))
		return s.ID
	}

	// countNotes は当該セッションのメモ行数。0 なら「1 行も作られていない」。
	countNotes := func(t *testing.T, sessionID uint64) int {
		t.Helper()
		var n int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT count(*) FROM session_notes WHERE session_id = $1`, sessionID).Scan(&n))
		return n
	}

	t.Run("メモがまだ無い他人のセッションには新規作成できない", func(t *testing.T) {
		sessionID := newSession(t)

		got, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: attacker, Content: "攻撃者が作成",
		})
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Nil(t, got)
		require.Equal(t, 0, countNotes(t, sessionID), "1 行も作られてはいけない")

		// 所有者がその後ふつうにメモを書けること（＝締め出されていないこと）まで確かめる。
		// 攻撃者名義の行ができていると、ここが 0 行 → not-found になって落ちる。
		saved, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: owner, Content: "所有者のメモ",
		})
		require.NoError(t, err)
		require.Equal(t, "所有者のメモ", saved.Content)

		read, err := get.Execute(ctx, usecase.GetSessionNoteInput{SessionID: sessionID, UserID: owner})
		require.NoError(t, err)
		require.NotNil(t, read, "所有者は自分のメモを読めなければならない")
		require.Equal(t, "所有者のメモ", read.Content)
	})

	t.Run("存在しないセッションには書けない", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "session_notes", "ai_chat_sessions")

		const missingSessionID = uint64(999_999)
		_, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: missingSessionID, UserID: owner, Content: "無いセッションへ",
		})
		// 他人のセッションのときと同じ not-found にする。応答を分けると、その差から
		// session_id が実在するかを総当たりで判別できてしまう。
		require.ErrorIs(t, err, domain.ErrNotFound)
		require.Equal(t, 0, countNotes(t, missingSessionID), "1 行も作られてはいけない")
	})

	t.Run("自分のセッションにはメモが無くても作成でき更新もできる", func(t *testing.T) {
		sessionID := newSession(t)

		created, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: owner, Content: "初回",
		})
		require.NoError(t, err)
		require.Equal(t, "初回", created.Content)

		updated, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: owner, Content: "更新後",
		})
		require.NoError(t, err)
		require.Equal(t, "更新後", updated.Content)
		require.Equal(t, created.ID, updated.ID, "同じ行が更新されること（行を増やさない）")
		require.Equal(t, 1, countNotes(t, sessionID))
	})

	t.Run("他人は既存メモを上書きできない", func(t *testing.T) {
		sessionID := newSession(t)
		_, err := upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: owner, Content: "所有者のメモ",
		})
		require.NoError(t, err)

		_, err = upsert.Execute(ctx, usecase.UpsertSessionNoteInput{
			SessionID: sessionID, UserID: attacker, Content: "攻撃者が上書き",
		})
		require.ErrorIs(t, err, domain.ErrNotFound)

		var uid uint64
		var content string
		require.NoError(t, db.QueryRowContext(
			ctx,
			`SELECT user_id, content FROM session_notes WHERE session_id = $1`, sessionID,
		).Scan(&uid, &content))
		require.Equal(t, owner, uid, "所有者が移ってはいけない")
		require.Equal(t, "所有者のメモ", content, "内容が変わってはいけない")
	})
}
