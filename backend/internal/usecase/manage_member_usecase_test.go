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

// manageStore は fake が記録する更新内容。
type manageStore struct {
	updateActiveGot *bool
	softDeleted     bool
}

// manageMemberRepo は UserRepository の fake を、このテストが使う 3 メソッドだけ
// 差し込んで返す。残り 8 メソッドは生成 fake がゼロ値を返すので no-op が要らない。
func manageMemberRepo(target *domain.User) (*repofakes.FakeUserRepository, *manageStore) {
	st := &manageStore{}
	repo := &repofakes.FakeUserRepository{
		FindByIDFunc: func(context.Context, uint64) (*domain.User, error) { return target, nil },
		UpdateActiveFunc: func(_ context.Context, _ uint64, active bool) error {
			st.updateActiveGot = &active
			return nil
		},
		SoftDeleteFunc: func(context.Context, uint64) error {
			st.softDeleted = true
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
