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
	db := testsupport.OpenTestDB(t)
	counter := persistence.NewCompanyStatsRepository(db)
	userRepo := persistence.NewUserRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "users", "companies")

	require.NoError(t, db.Create(&domain.Company{ID: 1, Name: "C1"}).Error)
	require.NoError(t, db.Create(&domain.Company{ID: 2, Name: "C2"}).Error)

	c1 := uint64(1)
	c2 := uint64(2)
	mk := func(sub string, cid *uint64, role string, active bool) *domain.User {
		return &domain.User{
			CognitoSub: sub, Email: sub + "@example.com", DisplayName: sub,
			Role: role, CompanyID: cid, IsActive: active,
		}
	}
	// 会社1: trainee有効 / trainee無効 / company_admin有効 / trainee(論理削除→除外)
	require.NoError(t, userRepo.Create(ctx, mk("a", &c1, domain.RoleTrainee, true)))
	require.NoError(t, userRepo.Create(ctx, mk("b", &c1, domain.RoleTrainee, false)))
	require.NoError(t, userRepo.Create(ctx, mk("c", &c1, domain.RoleCompanyAdmin, true)))
	require.NoError(t, userRepo.Create(ctx, mk("d", &c1, domain.RoleTrainee, true)))
	dUser, err := userRepo.FindByCognitoSub(ctx, "d")
	require.NoError(t, err)
	require.NoError(t, userRepo.SoftDelete(ctx, dUser.ID))
	// 会社2: trainee有効 1
	require.NoError(t, userRepo.Create(ctx, mk("e", &c2, domain.RoleTrainee, true)))
	// 会社未所属（super_admin）→ company_id IS NULL で除外
	require.NoError(t, userRepo.Create(ctx, mk("z", nil, domain.RoleSuperAdmin, true)))

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
