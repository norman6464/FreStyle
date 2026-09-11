//go:build integration

package user_test

import (
	"context"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
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
		uc := user.NewUpsertUserFromIDTokenUseCase(
			users,
			persistence.NewUserOidcIdentityRepository(db),
			persistence.NewTxManager(db),
		)

		got, err := uc.Execute(ctx, user.UpsertUserFromIDTokenInput{
			Subject:       "new-sub",
			Email:         "new@example.com",
			EmailVerified: true,
			Name:          "新規ユーザー",
		})
		require.NoError(t, err)
		require.NotNil(t, got)

		created, err := users.FindByOidcSubject(ctx, "new-sub")
		require.NoError(t, err)
		require.NotNil(t, created)
		require.Equal(t, "new@example.com", created.Email)
		require.Equal(t, "新規ユーザー", created.Name)
	})

	t.Run("既存ユーザーはidentityをセルフヒールし表示名を補完する", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		users := persistence.NewUserRepository(db)
		uc := user.NewUpsertUserFromIDTokenUseCase(
			users,
			persistence.NewUserOidcIdentityRepository(db),
			persistence.NewTxManager(db),
		)

		// 1 回目でユーザーを作る（Name は email と同じ = 未編集）。
		_, err := uc.Execute(ctx, user.UpsertUserFromIDTokenInput{
			Subject:       "existing-sub",
			Email:         "existing@example.com",
			EmailVerified: true,
		})
		require.NoError(t, err)

		// 2 回目、name claim 付きで再度ログイン。
		got, err := uc.Execute(ctx, user.UpsertUserFromIDTokenInput{
			Subject:       "existing-sub",
			Email:         "existing@example.com",
			EmailVerified: true,
			Name:          "後から付いた名前",
		})
		require.NoError(t, err)
		require.NotNil(t, got)

		got, err = users.FindByOidcSubject(ctx, "existing-sub")
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Equal(t, "後から付いた名前", got.Name, "未編集（Name==Email）なら OIDC name で補完される")
	})

	t.Run("未検証のサインアップはemailを保存せず後日検証済みで付く", func(t *testing.T) {
		testsupport.TruncateAll(t, db, "users", "user_oidc_identities")

		users := persistence.NewUserRepository(db)
		uc := user.NewUpsertUserFromIDTokenUseCase(
			users,
			persistence.NewUserOidcIdentityRepository(db),
			persistence.NewTxManager(db),
		)

		// 1 回目: サインアップ時点では未検証（GCIP のメール/パスワード登録の既定）。
		_, err := uc.Execute(ctx, user.UpsertUserFromIDTokenInput{
			Subject: "verify-later-sub",
			Email:   "victim@example.com",
			// EmailVerified は既定値 false のまま。
		})
		require.NoError(t, err)

		afterSignup, err := users.FindByOidcSubject(ctx, "verify-later-sub")
		require.NoError(t, err)
		require.NotNil(t, afterSignup)
		require.Equal(t, "", afterSignup.Email, "未検証の email は保存されないはず")

		// 2 回目: 発行者側で確認リンクを踏んだ後の再ログイン。
		_, err = uc.Execute(ctx, user.UpsertUserFromIDTokenInput{
			Subject:       "verify-later-sub",
			Email:         "victim@example.com",
			EmailVerified: true,
		})
		require.NoError(t, err)

		afterVerify, err := users.FindByOidcSubject(ctx, "verify-later-sub")
		require.NoError(t, err)
		require.NotNil(t, afterVerify)
		require.Equal(t, "victim@example.com", afterVerify.Email, "検証済みで再ログインしたら email が付くはず")
		require.Equal(t, afterSignup.ID, afterVerify.ID, "同じユーザー行のまま（新しい行を作らない）")
	})
}
