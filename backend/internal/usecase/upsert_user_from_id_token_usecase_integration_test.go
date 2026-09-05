//go:build integration

package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/require"
)

// TestUpsertUserFromIDToken_Integration は自己サインアップ（新規作成・既存ユーザーの
// identity セルフヒール）が実 PostgreSQL 上で users 行と user_oidc_identities を
// 不可分に作ることを固定する。
func TestUpsertUserFromIDToken_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("新規ユーザーは users 行と identity を対で作る", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		users := persistence.NewUserRepository(db)
		uc := usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			persistence.NewUserOidcIdentityRepository(db),
			persistence.NewTxManager(db),
		)

		user, err := uc.Execute(ctx, usecase.UpsertUserFromIDTokenInput{
			CognitoSub: "new-sub",
			Email:      "new@example.com",
			Name:       "新規ユーザー",
		})
		require.NoError(t, err)
		require.NotNil(t, user)

		created, err := users.FindByCognitoSub(ctx, "new-sub")
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, "new@example.com", created.Email)
		require.Equal(t, "新規ユーザー", created.Name)
	})

	t.Run("既存ユーザーはidentityをセルフヒールし表示名を補完する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		users := persistence.NewUserRepository(db)
		uc := usecase.NewUpsertUserFromIDTokenUseCase(
			users,
			persistence.NewUserOidcIdentityRepository(db),
			persistence.NewTxManager(db),
		)

		// 1 回目でユーザーを作る（Name は email と同じ = 未編集）。
		_, err := uc.Execute(ctx, usecase.UpsertUserFromIDTokenInput{
			CognitoSub: "existing-sub",
			Email:      "existing@example.com",
		})
		require.NoError(t, err)

		// 2 回目、name claim 付きで再度ログイン。
		user, err := uc.Execute(ctx, usecase.UpsertUserFromIDTokenInput{
			CognitoSub: "existing-sub",
			Email:      "existing@example.com",
			Name:       "後から付いた名前",
		})
		require.NoError(t, err)
		require.NotNil(t, user)

		got, err := users.FindByCognitoSub(ctx, "existing-sub")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "後から付いた名前", got.Name, "未編集（Name==Email）なら OIDC name で補完される")
	})
}
