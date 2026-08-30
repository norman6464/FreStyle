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

// TestCompanyStatsRepository_CountMembersByWorkspace_Integration は CountMembersByWorkspace の
// 集計（FILTER）と、論理削除済み / ワークスペース未所属の除外を実 Postgres で検証する。
func TestCompanyStatsRepository_CountMembersByWorkspace_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	counter := persistence.NewCompanyStatsRepository(sqlDB)
	userRepo := persistence.NewUserRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, sqlDB, tenantBridgeTables...)

	insertCompany(t, sqlDB, 1, "C1", true)
	insertCompany(t, sqlDB, 2, "C2", true)
	runStartupBackfill(ctx, t, sqlDB)
	ws1 := companyWorkspaceID(t, sqlDB, 1)
	require.True(t, ws1.Valid)
	ws2 := companyWorkspaceID(t, sqlDB, 2)
	require.True(t, ws2.Valid)

	ws1Str := ws1.UUID.String()
	ws2Str := ws2.UUID.String()
	// sub をそのまま OIDC subject に使い、users 行と identity を対で作る。
	create := func(sub string, workspaceID *string, role domain.RoleName) {
		u := &domain.User{
			Email: sub + "@example.com", Name: sub,
			Role: role, WorkspaceID: workspaceID, IsActive: true,
		}
		require.NoError(t, userRepo.CreateWithOidcIdentity(ctx, u, domain.OidcProviderCognito, sub))
	}
	// ワークスペース1: trainee有効 / trainee無効 / company_admin有効 / trainee(論理削除→除外)
	create("a", &ws1Str, domain.RoleTrainee)
	create("b", &ws1Str, domain.RoleTrainee)
	bUser, err := userRepo.FindByCognitoSub(ctx, "b")
	require.NoError(t, err)
	require.NoError(t, userRepo.UpdateActive(ctx, bUser.ID, false))
	create("c", &ws1Str, domain.RoleCompanyAdmin)
	create("d", &ws1Str, domain.RoleTrainee)
	dUser, err := userRepo.FindByCognitoSub(ctx, "d")
	require.NoError(t, err)
	require.NoError(t, userRepo.SoftDelete(ctx, dUser.ID))
	// ワークスペース2: trainee有効 1
	create("e", &ws2Str, domain.RoleTrainee)
	// ワークスペース未所属（super_admin）→ workspace_id が NULL で除外
	create("z", nil, domain.RoleSuperAdmin)

	rows, err := counter.CountMembersByWorkspace(ctx)
	require.NoError(t, err)

	byWorkspace := map[string]repository.WorkspaceMemberCount{}
	for _, r := range rows {
		byWorkspace[r.WorkspaceID] = r
	}

	// ワークスペース1: total 3（a,b,c。d は論理削除で除外）/ active 2（a,c）/ trainees 2（a,b）
	require.Equal(t, 3, byWorkspace[ws1Str].Total)
	require.Equal(t, 2, byWorkspace[ws1Str].Active)
	require.Equal(t, 2, byWorkspace[ws1Str].Trainees)
	// ワークスペース2: total 1 / active 1 / trainees 1
	require.Equal(t, 1, byWorkspace[ws2Str].Total)
	require.Equal(t, 1, byWorkspace[ws2Str].Active)
	require.Equal(t, 1, byWorkspace[ws2Str].Trainees)
	// ワークスペース未所属（workspace_id NULL）は集計に出ない
	require.Len(t, rows, 2)
}
