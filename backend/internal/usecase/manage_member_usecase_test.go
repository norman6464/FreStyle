package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// manageMemberRepo は FindByID が target を返す UserRepository の mock を返す。
// 更新系(UpdateActive / SoftDelete)の期待は各テストで明示し、
// AssertExpectations で「正しい引数で呼ばれたこと」まで検証する。
func manageMemberRepo(target *domain.User) *mockUserRepo {
	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, mock.Anything).Return(target, nil).Maybe()
	return repo
}

// otherWorkspaceID は memberWorkspaceID とは別のワークスペース（越境チェック用）。
const otherWorkspaceID = "0198a000-0000-7000-8000-000000000002"

func Test_メンバー有効化_会社管理者_自社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(memberWorkspaceID)})
	repo.On("UpdateActive", mock.Anything, uint64(2), false).Return(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	repo.AssertCalled(t, "FindByID", mock.Anything, uint64(2))
	repo.AssertExpectations(t) // UpdateActive(2, false) が呼ばれたこと
}

func Test_メンバー有効化_会社管理者_別ワークスペース_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(otherWorkspaceID)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_会社管理者_自身が未所属_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(memberWorkspaceID)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: nil}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_会社管理者_対象が未所属_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: nil})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_運営管理者_任意の会社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(memberWorkspaceID)})
	repo.On("UpdateActive", mock.Anything, uint64(2), false).Return(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_メンバー有効化_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_運営管理者_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}

	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_見つからない(t *testing.T) {
	repo := manageMemberRepo(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_会社管理者_自社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(memberWorkspaceID)})
	repo.On("SoftDelete", mock.Anything, uint64(2)).Return(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 2)

	require.NoError(t, err)
	repo.AssertCalled(t, "FindByID", mock.Anything, uint64(2))
	repo.AssertExpectations(t) // SoftDelete(2) が呼ばれたこと
}

func Test_メンバー論理削除_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 1)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_会社管理者_別ワークスペース_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(otherWorkspaceID)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, WorkspaceID: strptr(memberWorkspaceID)}

	err := uc.Execute(context.Background(), actor, 2)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_運営管理者_任意の会社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, WorkspaceID: strptr(otherWorkspaceID)})
	repo.On("SoftDelete", mock.Anything, uint64(2)).Return(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}

	err := uc.Execute(context.Background(), actor, 2)

	require.NoError(t, err)
	repo.AssertExpectations(t) // super_admin は会社を問わず削除できる
}

func Test_メンバー論理削除_運営管理者_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, WorkspaceID: nil}

	err := uc.Execute(context.Background(), actor, 1)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything) // super_admin でも自分自身は削除できない
}

func Test_メンバー論理削除_見つからない(t *testing.T) {
	repo := manageMemberRepo(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}
