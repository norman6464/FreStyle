//go:build integration

package persistence_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// insertUserWithRole は role_id を明示して users 行を 1 件作る。
// is_platform_admin 列が無い状態（バックフィル前）でも使えるよう、列を並べた生 INSERT で書く。
func insertUserWithRole(t *testing.T, db *gorm.DB, email string, roleID uint16) uint64 {
	t.Helper()
	var id uint64
	require.NoError(t, db.Raw(
		`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
		 VALUES (?, ?, ?, true, NOW(), NOW()) RETURNING id`,
		email, email, roleID,
	).Scan(&id).Error)
	return id
}

func platformAdminFlagOf(t *testing.T, db *gorm.DB, userID uint64) bool {
	t.Helper()
	var flag bool
	require.NoError(t, db.Raw(`SELECT is_platform_admin FROM users WHERE id = ?`, userID).Scan(&flag).Error)
	return flag
}

// TestExpandUsersPlatformAdmin_Integration は運営権限の受け皿の追加とバックフィルを実 PostgreSQL で固定する。
//
// 大事なのは 2 点:
//   - 適用直後は既存の super_admin が全員 true（＝挙動が変わらない）こと
//   - 何度流しても結果が変わらないこと。特に、false にした退任者が再適用で復権しないこと
//     （役割は下げないので「role が super_admin なら true」を毎起動流すと必ず復権する）
func TestExpandUsersPlatformAdmin_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, db, "user_oidc_identities", "users", "companies")

	// バックフィル前の状態を作る。OpenTestDB は本番と同じ順序で列を用意済みなので、
	// 一度落として「列が無い」ところからやり直す。
	require.NoError(t, db.Exec(`ALTER TABLE users DROP COLUMN IF EXISTS is_platform_admin`).Error)

	// ロールを増やしたときに初期データと期待値の更新漏れが起きないよう、1 つの表で持つ。
	cases := []struct {
		name   string
		email  string
		roleID uint16
		want   bool
	}{
		{"super_admin は運営権限を引き継ぐ", "expand-super@example.com", domain.RoleIDSuperAdmin, true},
		{"company_admin は対象外", "expand-company@example.com", domain.RoleIDCompanyAdmin, false},
		{"trainee は対象外", "expand-trainee@example.com", domain.RoleIDTrainee, false},
	}
	ids := make([]uint64, len(cases))
	for i, tc := range cases {
		ids[i] = insertUserWithRole(t, db, tc.email, tc.roleID)
	}

	// バックフィルは「列が無かった 1 回だけ」なので、全ロール分の行を作ってから一度だけ流す。
	// ケースごとに流し直すと 2 回目以降は列が既にあり no-op になり、何も検証できない。
	require.NoError(t, database.ExpandUsersPlatformAdmin(db))

	assertAll := func(t *testing.T) {
		t.Helper()
		for i, tc := range cases {
			require.Equal(t, tc.want, platformAdminFlagOf(t, db, ids[i]), tc.name)
		}
	}

	t.Run("適用直後はロールどおりの値になる", func(t *testing.T) {
		assertAll(t)
	})

	t.Run("再適用しても結果が変わらない", func(t *testing.T) {
		require.NoError(t, database.ExpandUsersPlatformAdmin(db))
		assertAll(t)
	})

	t.Run("失効させた運営管理者は再適用で復権しない", func(t *testing.T) {
		superAdmin := ids[0]
		require.NoError(t, db.Exec(
			`UPDATE users SET is_platform_admin = false WHERE id = ?`, superAdmin,
		).Error)

		require.NoError(t, database.ExpandUsersPlatformAdmin(db))
		require.False(t, platformAdminFlagOf(t, db, superAdmin),
			"role_id は下げないので、無条件のバックフィルを毎起動流すと退任者が復権する")

		// 後続のケースに引きずらないよう元へ戻す。
		require.NoError(t, db.Exec(
			`UPDATE users SET is_platform_admin = true WHERE id = ?`, superAdmin,
		).Error)
	})

	t.Run("バックフィル後に作られた super_admin は対象外", func(t *testing.T) {
		later := insertUserWithRole(t, db, "expand-later@example.com", domain.RoleIDSuperAdmin)
		require.False(t, platformAdminFlagOf(t, db, later),
			"列の既定値は false。付与はログイン時の同期か作成時の明示に限る")

		require.NoError(t, database.ExpandUsersPlatformAdmin(db))
		require.False(t, platformAdminFlagOf(t, db, later))
		assertAll(t)
	})
}
