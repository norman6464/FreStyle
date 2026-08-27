package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kbAdminA = "0198a000-0000-7000-8000-0000000000a1"
	kbAdminB = "0198a000-0000-7000-8000-0000000000a2"
	kbGroupA = "0198a000-0000-7000-8000-0000000000a3"
)

func kbUserPrincipal(id string, userID uint64) *domain.Principal {
	return &domain.Principal{
		ID: id, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser, UserID: &userID,
	}
}

func kbAdminGrant(principalID string) domain.WorkspaceGrant {
	return domain.WorkspaceGrant{WorkspaceID: kbWS, PrincipalID: principalID, Role: domain.GrantRoleAdmin}
}

func Test_最後のadmin判定_必須項目の検証(t *testing.T) {
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CanRemoveWorkspaceAdminInput{PrincipalID: kbAdminA})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CanRemoveWorkspaceAdminInput{WorkspaceID: kbWS})
	require.Error(t, err, "対象（principalID か userID）が必須")
}

func Test_最後のadmin判定_他にユーザーのadminが居れば外せる(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("ListWorkspaceGrants", mock.Anything, kbWS).
		Return([]domain.WorkspaceGrant{kbAdminGrant(kbAdminA), kbAdminGrant(kbAdminB)}, nil)
	repo.On("FindPrincipal", mock.Anything, kbWS, kbAdminB).Return(kbUserPrincipal(kbAdminB, 2), nil)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, PrincipalID: kbAdminA,
	})
	require.NoError(t, err)
	assert.True(t, ok)
}

func Test_最後のadmin判定_ひとりだけなら外せない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("ListWorkspaceGrants", mock.Anything, kbWS).
		Return([]domain.WorkspaceGrant{
			kbAdminGrant(kbAdminA),
			{WorkspaceID: kbWS, PrincipalID: kbAdminB, Role: domain.GrantRoleEditor},
		}, nil)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, PrincipalID: kbAdminA,
	})
	require.NoError(t, err)
	assert.False(t, ok, "admin が 0 人になる操作は断る")
}

func Test_最後のadmin判定_グループ宛てのadminは数えない(t *testing.T) {
	// メンバーが 1 人も居ないグループが「最後の admin」として残ると、結局誰も
	// 権限を変えられなくなる。grant の行からは中身が分からないので数に入れない。
	repo := &mockKBPermissionRepo{}
	repo.On("ListWorkspaceGrants", mock.Anything, kbWS).
		Return([]domain.WorkspaceGrant{kbAdminGrant(kbAdminA), kbAdminGrant(kbGroupA)}, nil)
	repo.On("FindPrincipal", mock.Anything, kbWS, kbGroupA).
		Return(&domain.Principal{ID: kbGroupA, WorkspaceID: kbWS, Kind: domain.PrincipalKindGroup, Name: "運用"}, nil)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, PrincipalID: kbAdminA,
	})
	require.NoError(t, err)
	assert.False(t, ok)
}

func Test_最後のadmin判定_元からadminでなければ通す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("ListWorkspaceGrants", mock.Anything, kbWS).
		Return([]domain.WorkspaceGrant{
			{WorkspaceID: kbWS, PrincipalID: kbAdminA, Role: domain.GrantRoleViewer},
		}, nil)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, PrincipalID: kbAdminA,
	})
	require.NoError(t, err)
	assert.True(t, ok, "admin を 1 人も減らさない操作は止めない")
}

func Test_最後のadmin判定_ユーザーIDで指しても同じ結論になる(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(1)).Return(kbUserPrincipal(kbAdminA, 1), nil)
	repo.On("ListWorkspaceGrants", mock.Anything, kbWS).
		Return([]domain.WorkspaceGrant{kbAdminGrant(kbAdminA)}, nil)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, UserID: 1,
	})
	require.NoError(t, err)
	assert.False(t, ok, "メンバー削除でも principal ごと消えて admin が 0 人になる")
}

func Test_最後のadmin判定_非メンバーは通す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(9)).
		Return(nil, repository.ErrPrincipalNotFound)
	uc := usecase.NewCanRemoveWorkspaceAdminUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: kbWS, UserID: 9,
	})
	require.NoError(t, err)
	assert.True(t, ok)
}
