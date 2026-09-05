//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// workspaceWriteTables はこの節のテストが TRUNCATE する対象。
var workspaceWriteTables = []string{"users", "user_oidc_identities", "workspaces"}

// insertWorkspaceWithActive はワークスペースを 1 件作る（id を明示して固定する）。
func insertWorkspaceWithActive(t *testing.T, db *sql.DB, id uuid.UUID, name string, active bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO workspaces (id, slug, name, is_active) VALUES ($1, $2, $3, $4)`,
		id, "ws-"+id.String(), name, active,
	)
	require.NoError(t, err)
}

// userWorkspaceID はユーザーの workspace_id を返す（未設定なら Valid=false）。
func userWorkspaceID(t *testing.T, db *sql.DB, userID uint64) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(`SELECT workspace_id FROM users WHERE id = $1`, userID).Scan(&got))
	return got
}

// createUserWithOidcIdentity は users 行と identity をフィクスチャとして用意する。
// usecase.UpsertUserFromIDTokenUseCase の新規作成経路と同じ2段（Create → EnsureIdentity）。
// ここではロールバックの検証はしないため、DoInTx で束ねずそのまま呼ぶ。
func createUserWithOidcIdentity(ctx context.Context, t *testing.T, db *sql.DB, user *domain.User, provider, subject string) {
	t.Helper()
	require.NoError(t, persistence.NewUserRepository(db).Create(ctx, user))
	require.NoError(t, persistence.NewUserOidcIdentityRepository(db).EnsureIdentity(ctx, user.ID, provider, subject))
}

// TestUserWorkspaceWrite_Integration は users.workspace_id が「渡された値そのまま」で
// 書かれることを固定する。所属参照の解決は呼び出し側（usecase）の責務で、
// リポジトリ側が引き直したりはしない。
func TestUserWorkspaceWrite_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("渡した所属ワークスペースがそのまま書かれる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1, ws2 := uuid.New(), uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)
		ws1Str, ws2Str := ws1.String(), ws2.String()

		repo := persistence.NewUserRepository(sqlDB)
		createUserWithOidcIdentity(ctx, t, sqlDB, &domain.User{
			Email: "a@example.com", Name: "A", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-a")
		createUserWithOidcIdentity(ctx, t, sqlDB, &domain.User{
			Email: "b@example.com", Name: "B", WorkspaceID: &ws2Str,
		}, domain.OidcProviderCognito, "sub-b")

		userA, err := repo.FindByCognitoSub(ctx, "sub-a")
		require.NoError(t, err)
		userB, err := repo.FindByCognitoSub(ctx, "sub-b")
		require.NoError(t, err)

		require.Equal(t, uuid.NullUUID{UUID: ws1, Valid: true}, userWorkspaceID(t, sqlDB, userA.ID))
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, userWorkspaceID(t, sqlDB, userB.ID))
		// 読み出しも同じ所属を返す。
		require.Equal(t, ws1Str, *userA.WorkspaceID)
		require.Equal(t, ws2Str, *userB.WorkspaceID)
	})

	t.Run("未所属のまま作られたユーザーは所属先ワークスペースが無い", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		repo := persistence.NewUserRepository(sqlDB)
		createUserWithOidcIdentity(ctx, t, sqlDB, &domain.User{
			Email: "root@example.com", Name: "運営",
		}, domain.OidcProviderCognito, "sub-root")

		got, err := repo.FindByCognitoSub(ctx, "sub-root")
		require.NoError(t, err)
		require.Nil(t, got.WorkspaceID)
		require.False(t, tableWorkspaceID(t, sqlDB, "users", got.ID).Valid)
	})

	t.Run("所属の付け替えでワークスペースが入れ替わる", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1, ws2 := uuid.New(), uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		insertWorkspaceWithActive(t, sqlDB, ws2, "ワークスペース B", true)
		ws1Str, ws2Str := ws1.String(), ws2.String()

		repo := persistence.NewUserRepository(sqlDB)
		createUserWithOidcIdentity(ctx, t, sqlDB, &domain.User{
			Email: "move@example.com", Name: "異動", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-move")
		got, err := repo.FindByCognitoSub(ctx, "sub-move")
		require.NoError(t, err)
		require.Equal(t, uuid.NullUUID{UUID: ws1, Valid: true}, tableWorkspaceID(t, sqlDB, "users", got.ID))

		require.NoError(t, repo.UpdateWorkspaceID(ctx, got.ID, &ws2Str))

		moved, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Equal(t, ws2Str, *moved.WorkspaceID)
		require.Equal(t, uuid.NullUUID{UUID: ws2, Valid: true}, tableWorkspaceID(t, sqlDB, "users", got.ID))
	})

	t.Run("nil を渡すと未所属へ戻る", func(t *testing.T) {
		testsupport.TruncateAll(t, sqlDB, workspaceWriteTables...)
		ws1 := uuid.New()
		insertWorkspaceWithActive(t, sqlDB, ws1, "ワークスペース A", true)
		ws1Str := ws1.String()

		repo := persistence.NewUserRepository(sqlDB)
		createUserWithOidcIdentity(ctx, t, sqlDB, &domain.User{
			Email: "leave@example.com", Name: "退所", WorkspaceID: &ws1Str,
		}, domain.OidcProviderCognito, "sub-leave")
		got, err := repo.FindByCognitoSub(ctx, "sub-leave")
		require.NoError(t, err)

		require.NoError(t, repo.UpdateWorkspaceID(ctx, got.ID, nil))

		left, err := repo.FindByID(ctx, got.ID)
		require.NoError(t, err)
		require.Nil(t, left.WorkspaceID)
		require.False(t, userWorkspaceID(t, sqlDB, got.ID).Valid)
	})
}

// tableWorkspaceID は業務テーブル 1 行の workspace_id を返す。table は本ファイル内の
// ハードコードされた定数のみを渡す（外部入力を SQL に組み込まない）。
func tableWorkspaceID(t *testing.T, db *sql.DB, table string, id any) uuid.NullUUID {
	t.Helper()
	var got uuid.NullUUID
	require.NoError(t, db.QueryRow(fmt.Sprintf(`SELECT workspace_id FROM %s WHERE id = $1`, table), id).Scan(&got))
	return got
}
