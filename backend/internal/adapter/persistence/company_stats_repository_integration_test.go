//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/require"
)

// TestCompanyStatsRepository_Integration は CountMembersByCompany の集計（FILTER）と、
// 論理削除済み / 会社未所属の除外を実 Postgres で検証する。
func TestCompanyStatsRepository_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	counter := persistence.NewCompanyStatsRepository(sqlDB)
	userRepo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, "users", "companies")

	for _, c := range []struct {
		id   uint64
		name string
	}{{1, "C1"}, {2, "C2"}} {
		_, err := sqlDB.Exec(
			`INSERT INTO companies (id, name, created_at, updated_at) VALUES ($1, $2, now(), now())`,
			c.id, c.name,
		)
		require.NoError(t, err)
	}

	c1 := uint64(1)
	c2 := uint64(2)
	// sub をそのまま OIDC subject に使い、users 行と identity を対で作る。
	create := func(sub string, cid *uint64, role domain.RoleName) {
		u := &domain.User{
			Email: sub + "@example.com", Name: sub,
			Role: role, CompanyID: cid, IsActive: true,
		}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
	}
	// 会社1: trainee有効 / trainee無効 / company_admin有効 / trainee(論理削除→除外)
	create("a", &c1, domain.RoleTrainee)
	// b は無効ユーザー。GORM は is_active の default:true 指定によりゼロ値(false)を省略して
	// DB デフォルト true で挿入するため、作成後に明示的に UpdateActive(false) で無効化する。
	create("b", &c1, domain.RoleTrainee)
	bUser, err := userRepo.FindByCognitoSub(ctx, "b")
	require.NoError(t, err)
	require.NoError(t, userRepo.UpdateActive(ctx, bUser.ID, false))
	create("c", &c1, domain.RoleCompanyAdmin)
	create("d", &c1, domain.RoleTrainee)
	dUser, err := userRepo.FindByCognitoSub(ctx, "d")
	require.NoError(t, err)
	require.NoError(t, userRepo.SoftDelete(ctx, dUser.ID))
	// 会社2: trainee有効 1
	create("e", &c2, domain.RoleTrainee)
	// 会社未所属（super_admin）→ company_id IS NULL で除外
	create("z", nil, domain.RoleSuperAdmin)

	rows, err := counter.CountMembersByCompany(ctx)
	require.NoError(t, err)

	byID := map[uint64]repository.CompanyMemberCount{}
	for _, r := range rows {
		byID[r.CompanyID] = r
	}

	// 会社1: total 3（a,b,c。d は論理削除で除外）/ active 2（a,c）/ trainees 2（a,b）
	require.Equal(t, 3, byID[1].Total)
	require.Equal(t, 2, byID[1].Active)
	require.Equal(t, 2, byID[1].Trainees)
	// 会社2: total 1 / active 1 / trainees 1
	require.Equal(t, 1, byID[2].Total)
	require.Equal(t, 1, byID[2].Active)
	require.Equal(t, 1, byID[2].Trainees)
	// 会社未所属（company_id NULL）は集計に出ない
	_, ok := byID[0]
	require.False(t, ok)
}
