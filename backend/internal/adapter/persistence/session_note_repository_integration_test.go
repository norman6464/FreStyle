//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestSessionNoteRepository_Integration は FindBySessionID（round-trip / not-found）を実 Postgres で検証する。
func TestSessionNoteRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(sqlDB)
	ctx := context.Background()

	t.Run("Upsert → FindBySessionID で round-trip（1 行を返す）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "session_notes")
		require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{SessionID: 55, UserID: 7, Content: "メモ本文"}))

		got, err := repo.FindBySessionID(ctx, 55)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, uint64(55), got.SessionID)
		require.Equal(t, uint64(7), got.UserID)
		require.Equal(t, "メモ本文", got.Content)
	})

	t.Run("未作成は (nil, nil)", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, "session_notes")
		got, err := repo.FindBySessionID(ctx, 999)
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

// TestSessionNoteRepository_Upsert_WriteBack_Integration は Upsert が id / created_at /
// updated_at を呼び出し元の struct へ書き戻すことを実 Postgres で固定する。
func TestSessionNoteRepository_Upsert_WriteBack_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "session_notes")

	n := &domain.SessionNote{SessionID: 77, UserID: 3, Content: "本文"}
	require.NoError(t, repo.Upsert(ctx, n))
	require.NotZero(t, n.ID, "id が採番され書き戻される")
	require.False(t, n.CreatedAt.IsZero(), "created_at が書き戻される")
	require.False(t, n.UpdatedAt.IsZero(), "updated_at が書き戻される")

	got, err := repo.FindBySessionID(ctx, 77)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, n.ID, got.ID)
	require.Equal(t, "本文", got.Content)
}

// TestSessionNoteRepository_Upsert_SecondSaveUpdates_Integration は同じ session_id への
// 2 回目の保存が UPDATE になり行が増えないことを実 Postgres で固定する。
// content は新しい値・created_at は初回のまま・updated_at は進む・id も変わらない。
func TestSessionNoteRepository_Upsert_SecondSaveUpdates_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "session_notes")

	require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{SessionID: 88, UserID: 5, Content: "初回"}))
	first, err := repo.FindBySessionID(ctx, 88)
	require.NoError(t, err)
	require.NotNil(t, first)

	// now() はトランザクション開始時刻。2 回目が別トランザクション（別 now()）になるよう少し空ける。
	time.Sleep(5 * time.Millisecond)

	require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{SessionID: 88, UserID: 5, Content: "更新後"}))

	// 行が増えていない（session_id=88 も全体も 1 行）。
	var bySession, total int
	require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM session_notes WHERE session_id = 88`).Scan(&bySession))
	require.Equal(t, 1, bySession, "同じ session_id で行が増えてはいけない（UPDATE のはず）")
	require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM session_notes`).Scan(&total))
	require.Equal(t, 1, total)

	second, err := repo.FindBySessionID(ctx, 88)
	require.NoError(t, err)
	require.NotNil(t, second)
	require.Equal(t, first.ID, second.ID, "id は据え置き（同じ行の UPDATE）")
	require.Equal(t, "更新後", second.Content, "content は新しい値")
	require.Equal(t, uint64(5), second.UserID, "user_id は保持")
	require.WithinDuration(t, first.CreatedAt, second.CreatedAt, 0, "created_at は初回のまま")
	require.True(t, second.UpdatedAt.After(first.UpdatedAt), "updated_at は進む")
}

// TestSessionNoteRepository_UniqueConstraint_Integration は session_id の一意制約が実際に
// 効いていること（ON CONFLICT を経由しない直接 INSERT の 2 行目が 23505 で弾かれること）を固定する。
func TestSessionNoteRepository_UniqueConstraint_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewSessionNoteRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "session_notes")

	require.NoError(t, repo.Upsert(ctx, &domain.SessionNote{SessionID: 91, UserID: 5, Content: "1 行目"}))

	// upsert を経由せず生 INSERT で同じ session_id の 2 行目を作ろうとすると一意制約が弾く。
	_, err := sqlDB.Exec(
		`INSERT INTO session_notes (session_id, user_id, content, created_at, updated_at)
		 VALUES ($1, $2, $3, now(), now())`,
		91, 5, "2 行目",
	)
	requireSQLState(t, err, sqlStateUniqueViolation)
}

// TestSessionNoteRepository_UniqueIndexesPresent_Integration はスキーマ DDL 経路
// （schema/core.sql の idx_session_notes_session_id）と Apply 経路
// （ApplySessionNoteConstraints の明示 SQL）の両方で、テスト DB の
// session_notes(session_id) に一意インデックスが張られていることを固定する。
func TestSessionNoteRepository_UniqueIndexesPresent_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	names := sessionIDUniqueIndexNames(t, sqlDB)

	// Apply 経路: 明示 SQL が張る名前付きインデックス。
	require.Contains(t, names, "uq_session_notes_session_id", "明示 SQL の一意インデックスが無い（Apply 経路）")

	// スキーマ DDL 経路: CREATE TABLE と一緒に張る別名の一意インデックス。
	fromSchemaDDL := false
	for _, n := range names {
		if n != "uq_session_notes_session_id" {
			fromSchemaDDL = true
		}
	}
	require.Truef(t, fromSchemaDDL, "スキーマ DDL 経路の一意インデックスが無い（見つかった索引: %v）", names)
}

// sessionIDUniqueIndexNames は session_notes(session_id) を覆う一意インデックスの名前を返す。
func sessionIDUniqueIndexNames(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query(
		`SELECT indexname FROM pg_indexes
		 WHERE tablename = 'session_notes'
		   AND indexdef ILIKE '%UNIQUE INDEX%'
		   AND indexdef ILIKE '%(session_id)%'`,
	)
	require.NoError(t, err)
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())
	return names
}
