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

// userTxTables は users まわりの書き込み経路が触るテーブル。
var userTxTables = []string{"users", "user_oidc_identities", "workspaces"}

// userUpdatedAt はユーザーの updated_at を DB から直接読む。
func userUpdatedAt(t *testing.T, db *sql.DB, id uint64) time.Time {
	t.Helper()
	var ts time.Time
	require.NoError(t, db.QueryRow(`SELECT updated_at FROM users WHERE id = $1`, id).Scan(&ts))
	return ts
}

// TestUserRepositoryWrites_Integration は users の書き込み経路（オフボーディングの芯である
// UpdateActive / SoftDelete を含む）の契約を実 PostgreSQL で固定する。
func TestUserRepositoryWrites_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	repo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()

	// newTrainee は研修生を 1 人作って返す。
	newTrainee := func(t *testing.T, email, sub string) *domain.User {
		t.Helper()
		u := &domain.User{Email: email, Name: email}
		require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
		require.NotZero(t, u.ID, "作成後は採番された id が書き戻る")
		require.False(t, u.CreatedAt.IsZero(), "作成後は created_at が書き戻る")
		require.False(t, u.UpdatedAt.IsZero(), "作成後は updated_at が書き戻る")
		return u
	}

	t.Run("作成直後は有効（is_active=true）", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "active@example.com", "active-1")
		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.True(t, got.IsActive)
	})

	t.Run("UpdateActive(false) は即時に効き FindActiveByEmail から消える", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "off@example.com", "off-1")
		before := userUpdatedAt(t, sqlDB, u.ID)

		require.NoError(t, repo.UpdateActive(ctx, u.ID, false))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.False(t, got.IsActive)
		byEmail, err := repo.FindActiveByEmail(ctx, "off@example.com")
		require.NoError(t, err)
		require.Nil(t, byEmail, "無効化されたユーザーはログイン経路から引けない")
		require.NotEqual(t, before, userUpdatedAt(t, sqlDB, u.ID), "updated_at が進む")

		// 戻せる。
		require.NoError(t, repo.UpdateActive(ctx, u.ID, true))
		byEmail, err = repo.FindActiveByEmail(ctx, "off@example.com")
		require.NoError(t, err)
		require.NotNil(t, byEmail)
	})

	t.Run("UpdateActive は存在しないユーザーで domain.ErrNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		require.ErrorIs(t, repo.UpdateActive(ctx, 999999, false), domain.ErrNotFound)
	})

	t.Run("SoftDelete は identity を解放し、二度目は domain.ErrNotFound", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "bye@example.com", "bye-1")

		require.NoError(t, repo.SoftDelete(ctx, u.ID))

		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Nil(t, got, "論理削除後は引けない")
		bySub, err := repo.FindByCognitoSub(ctx, "bye-1")
		require.NoError(t, err)
		require.Nil(t, bySub)
		var n int64
		require.NoError(t, sqlDB.QueryRow(`SELECT count(*) FROM user_oidc_identities WHERE user_id = $1`, u.ID).Scan(&n))
		require.Equal(t, int64(0), n, "identity は解放される")

		require.ErrorIs(t, repo.SoftDelete(ctx, u.ID), domain.ErrNotFound)
		require.ErrorIs(t, repo.SoftDelete(ctx, 999999), domain.ErrNotFound)
	})

	t.Run("UpdateName は指定列だけを更新する", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "rename@example.com", "rename-1")
		before := userUpdatedAt(t, sqlDB, u.ID)

		require.NoError(t, repo.UpdateName(ctx, u.ID, "新しい名前"))
		got, err := repo.FindByID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, "新しい名前", got.Name)
		require.Equal(t, "rename@example.com", got.Email, "email は触らない")
		require.True(t, got.IsActive, "is_active は触らない")
		require.NotEqual(t, before, userUpdatedAt(t, sqlDB, u.ID))
	})

	t.Run("ListByWorkspaceID はワークスペースで絞り id 昇順・論理削除を除く", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		ws1 := createWorkspace(t, sqlDB, "list-by-workspace-a")
		ws2 := createWorkspace(t, sqlDB, "list-by-workspace-b")
		a := &domain.User{Email: "m1@example.com", Name: "m1", WorkspaceID: &ws1}
		b := &domain.User{Email: "m2@example.com", Name: "m2", WorkspaceID: &ws1}
		other := &domain.User{Email: "m3@example.com", Name: "m3", WorkspaceID: &ws2}
		for _, u := range []*domain.User{a, b, other} {
			require.NoError(t, repo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, u.Name))
		}

		rows, err := repo.ListByWorkspaceID(ctx, ws1)
		require.NoError(t, err)
		require.Len(t, rows, 2)
		require.Equal(t, "m1", rows[0].Name)
		require.Equal(t, "m2", rows[1].Name)

		require.NoError(t, repo.SoftDelete(ctx, b.ID))
		rows, err = repo.ListByWorkspaceID(ctx, ws1)
		require.NoError(t, err)
		require.Len(t, rows, 1)

		// 該当なしでも nil ではなく空スライス（JSON が null にならない）。
		empty, err := repo.ListByWorkspaceID(ctx, "0198a000-0000-7000-8000-0000000000ff")
		require.NoError(t, err)
		require.NotNil(t, empty)
		require.Empty(t, empty)

		// 不正な文字列（uuid として解釈できない）も該当なしと同じ扱い。
		invalid, err := repo.ListByWorkspaceID(ctx, "not-a-uuid")
		require.NoError(t, err)
		require.NotNil(t, invalid)
		require.Empty(t, invalid)
	})

	t.Run("CognitoSubjectByUserID は subject を返し、無ければ空文字", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		u := newTrainee(t, "sub@example.com", "sub-1")

		got, err := repo.CognitoSubjectByUserID(ctx, u.ID)
		require.NoError(t, err)
		require.Equal(t, "sub-1", got)

		got, err = repo.CognitoSubjectByUserID(ctx, 999999)
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("EnsureOidcIdentity は他人が持つ subject を奪わない", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, userTxTables...)
		owner := newTrainee(t, "owner@example.com", "shared-subject")
		other := newTrainee(t, "other@example.com", "other-subject")

		// 自分の subject を張り直しても冪等に成功する。
		require.NoError(t, repo.EnsureOidcIdentity(ctx, owner.ID, domain.OidcProviderCognito, "shared-subject"))

		// 他人が持つ subject を要求したら黙って成功にせずエラー。
		err := repo.EnsureOidcIdentity(ctx, other.ID, domain.OidcProviderCognito, "shared-subject")
		require.ErrorContains(t, err, "oidc identity conflict")

		// 持ち主は変わっていない。
		got, err := repo.FindByCognitoSub(ctx, "shared-subject")
		require.NoError(t, err)
		require.Equal(t, owner.ID, got.ID)
	})
}
