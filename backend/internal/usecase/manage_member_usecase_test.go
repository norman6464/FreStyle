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

func Test_メンバー有効化_会社管理者_自社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	repo.On("UpdateActive", mock.Anything, uint64(2), false).Return(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	repo.AssertCalled(t, "FindByID", mock.Anything, uint64(2))
	repo.AssertExpectations(t) // UpdateActive(2, false) が呼ばれたこと
}

func Test_メンバー有効化_会社管理者_別会社_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_運営管理者_任意の会社_OK(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	repo.On("UpdateActive", mock.Anything, uint64(2), false).Return(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func Test_メンバー有効化_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "UpdateActive", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー有効化_運営管理者_自分自身_禁止(t *testing.T){
	repo := &fakeManageRepo{target: &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}
	
	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	assert.Nil(t, repo.updateActiveGot, "super_admin でも自分自身は無効かできない")
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
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	repo.On("SoftDelete", mock.Anything, uint64(2)).Return(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.NoError(t, err)
	repo.AssertCalled(t, "FindByID", mock.Anything, uint64(2))
	repo.AssertExpectations(t) // SoftDelete(2) が呼ばれたこと
}

func Test_メンバー論理削除_自分自身_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_会社管理者_別会社_禁止(t *testing.T) {
	repo := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_見つからない(t *testing.T) {
	repo := manageMemberRepo(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
	repo.AssertNotCalled(t, "SoftDelete", mock.Anything, mock.Anything)
}

func Test_メンバー論理削除_運営管理者_任意の会社_OK(t *testing.T) {
      repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)}}
      uc := usecase.NewSoftDeleteMemberUseCase(repo)
      actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}

      err := uc.Execute(context.Background(), actor, 2)

      require.NoError(t, err)
      assert.True(t, repo.softDeleted, "super_admin は会社を問わず削除できる")
}

func Test_メンバー論理削除_運営管理者_自分自身_禁止(t *testing.T) {
      repo := &fakeManageRepo{target: &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}}
      uc := usecase.NewSoftDeleteMemberUseCase(repo)
      actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}

      err := uc.Execute(context.Background(), actor, 1)

      require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
      assert.False(t, repo.softDeleted, "super_admin でも自分自身は削除できない")
}

func Test_メンバー論理削除_見つからない(t *testing.T) {
      repo := &fakeManageRepo{target: nil}
      uc := usecase.NewSoftDeleteMemberUseCase(repo)
      actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

      err := uc.Execute(context.Background(), actor, 999)

      require.ErrorIs(t, err, usecase.ErrMemberNotFound)
}
