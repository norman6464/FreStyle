package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository/repofakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// manageStore は fake が記録する更新内容。対象 ID も控え、別のユーザーに
// 操作が飛んでいないことまで検証できるようにする。
type manageStore struct {
	updateActiveGot    *bool
	updateActiveUserID uint64
	softDeleted        bool
	softDeletedUserID  uint64
}

// manageMemberRepo は UserRepository の fake を、このテストが使う 3 メソッドだけ
// 差し込んで返す。残り 8 メソッドは生成 fake がゼロ値を返すので no-op が要らない。
func manageMemberRepo(target *domain.User) (*repofakes.FakeUserRepository, *manageStore) {
	st := &manageStore{}
	repo := &repofakes.FakeUserRepository{
		FindByIDFunc: func(context.Context, uint64) (*domain.User, error) { return target, nil },
		UpdateActiveFunc: func(_ context.Context, userID uint64, active bool) error {
			st.updateActiveGot = &active
			st.updateActiveUserID = userID
			return nil
		},
		SoftDeleteFunc: func(_ context.Context, userID uint64) error {
			st.softDeleted = true
			st.softDeletedUserID = userID
			return nil
		},
	}
	return repo, st
}

func Test_メンバー有効化_会社管理者_自社_OK(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	require.NotNil(t, store.updateActiveGot)
	assert.False(t, *store.updateActiveGot)
	assert.Equal(t, uint64(2), store.updateActiveUserID, "指定した対象に対して更新すること")
}

func Test_メンバー有効化_会社管理者_別会社_禁止(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	assert.Nil(t, store.updateActiveGot, "別会社では更新してはならない")
}

func Test_メンバー有効化_運営管理者_任意の会社_OK(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	require.NotNil(t, store.updateActiveGot)
	assert.False(t, *store.updateActiveGot, "実装が true を渡す回帰を検出する")
	assert.Equal(t, uint64(2), store.updateActiveUserID)
}

func Test_メンバー有効化_自分自身_禁止(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)})
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	assert.Nil(t, store.updateActiveGot, "自分自身は無効化できない")
}

func Test_メンバー有効化_見つからない(t *testing.T) {
	repo, _ := manageMemberRepo(nil)
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
	assert.Zero(t, repo.UpdateActiveCalls.Load(), "対象が居なければ更新は走らない")
}

func Test_メンバー論理削除_会社管理者_自社_OK(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.NoError(t, err)
	assert.True(t, store.softDeleted)
	assert.Equal(t, uint64(2), store.softDeletedUserID, "指定した対象を削除すること")
}

func Test_メンバー論理削除_自分自身_禁止(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	assert.False(t, store.softDeleted, "自分自身は削除できない")
}

func Test_メンバー論理削除_会社管理者_別会社_禁止(t *testing.T) {
	repo, store := manageMemberRepo(&domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)})
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	assert.False(t, store.softDeleted)
}

func Test_メンバー論理削除_見つからない(t *testing.T) {
	repo, _ := manageMemberRepo(nil)
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
	assert.Zero(t, repo.SoftDeleteCalls.Load(), "対象が居なければ削除は走らない")
}
