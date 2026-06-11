package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeManageRepo は SetMemberActive / SoftDeleteMember の認可検証用 fake。
type fakeManageRepo struct {
	target          *domain.User // FindByID が返す対象
	updateActiveGot *bool
	softDeleted     bool
}

func (f *fakeManageRepo) FindByID(context.Context, uint64) (*domain.User, error) {
	return f.target, nil
}

func (f *fakeManageRepo) UpdateActive(_ context.Context, _ uint64, active bool) error {
	f.updateActiveGot = &active
	return nil
}

func (f *fakeManageRepo) SoftDelete(context.Context, uint64) error {
	f.softDeleted = true
	return nil
}

func (f *fakeManageRepo) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeManageRepo) ListByRole(context.Context, string) ([]domain.User, error) { return nil, nil }

func (f *fakeManageRepo) ListByCompanyID(context.Context, uint64) ([]domain.User, error) {
	return nil, nil
}

func (f *fakeManageRepo) Create(context.Context, *domain.User) error { return nil }

func (f *fakeManageRepo) UpdateAiChatEnabled(context.Context, uint64, *bool) error { return nil }

func (f *fakeManageRepo) UpdateDisplayName(context.Context, uint64, string) error { return nil }

func (f *fakeManageRepo) UpdateRole(context.Context, uint64, string) error { return nil }

func (f *fakeManageRepo) UpdateCompanyID(context.Context, uint64, uint64) error { return nil }

func (f *fakeManageRepo) MarkOnboarded(context.Context, uint64) error { return nil }

func TestSetMemberActive_CompanyAdmin_OwnCompany_OK(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)}}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	require.NotNil(t, repo.updateActiveGot)
	assert.False(t, *repo.updateActiveGot)
}

func TestSetMemberActive_CompanyAdmin_OtherCompany_Forbidden(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)}}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	assert.Nil(t, repo.updateActiveGot, "別会社では更新してはならない")
}

func TestSetMemberActive_SuperAdmin_AnyCompany_OK(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)}}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin, CompanyID: nil}

	err := uc.Execute(context.Background(), actor, 2, false)

	require.NoError(t, err)
	require.NotNil(t, repo.updateActiveGot)
}

func TestSetMemberActive_Self_Forbidden(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1, false)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	assert.Nil(t, repo.updateActiveGot, "自分自身は無効化できない")
}

func TestSetMemberActive_NotFound(t *testing.T) {
	repo := &fakeManageRepo{target: nil}
	uc := usecase.NewSetMemberActiveUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleSuperAdmin}

	err := uc.Execute(context.Background(), actor, 999, false)

	require.ErrorIs(t, err, usecase.ErrMemberNotFound)
}

func TestSoftDeleteMember_CompanyAdmin_OwnCompany_OK(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(10)}}
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.NoError(t, err)
	assert.True(t, repo.softDeleted)
}

func TestSoftDeleteMember_Self_Forbidden(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}}
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 1)

	require.ErrorIs(t, err, usecase.ErrCannotManageSelf)
	assert.False(t, repo.softDeleted, "自分自身は削除できない")
}

func TestSoftDeleteMember_CompanyAdmin_OtherCompany_Forbidden(t *testing.T) {
	repo := &fakeManageRepo{target: &domain.User{ID: 2, Role: domain.RoleTrainee, CompanyID: u64ptr(99)}}
	uc := usecase.NewSoftDeleteMemberUseCase(repo)
	actor := &domain.User{ID: 1, Role: domain.RoleCompanyAdmin, CompanyID: u64ptr(10)}

	err := uc.Execute(context.Background(), actor, 2)

	require.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	assert.False(t, repo.softDeleted)
}
