package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMemberUserRepo は UserRepository を満たすテスト用 fake。
type fakeMemberUserRepo struct {
	byCompany map[uint64][]domain.User
	byID      map[uint64]*domain.User
	updated   map[uint64]*bool
}

func newFakeMemberUserRepo() *fakeMemberUserRepo {
	return &fakeMemberUserRepo{
		byCompany: map[uint64][]domain.User{},
		byID:      map[uint64]*domain.User{},
		updated:   map[uint64]*bool{},
	}
}

func (f *fakeMemberUserRepo) FindByCognitoSub(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (f *fakeMemberUserRepo) FindByID(_ context.Context, id uint64) (*domain.User, error) {
	return f.byID[id], nil
}

func (f *fakeMemberUserRepo) ListByRole(context.Context, string) ([]domain.User, error) {
	return nil, nil
}

func (f *fakeMemberUserRepo) ListByCompanyID(_ context.Context, companyID uint64) ([]domain.User, error) {
	return f.byCompany[companyID], nil
}
func (f *fakeMemberUserRepo) Create(context.Context, *domain.User) error { return nil }
func (f *fakeMemberUserRepo) UpdateDisplayName(context.Context, uint64, string) error {
	return nil
}
func (f *fakeMemberUserRepo) UpdateRole(context.Context, uint64, string) error      { return nil }
func (f *fakeMemberUserRepo) UpdateCompanyID(context.Context, uint64, uint64) error { return nil }
func (f *fakeMemberUserRepo) UpdateActive(context.Context, uint64, bool) error      { return nil }
func (f *fakeMemberUserRepo) SoftDelete(context.Context, uint64) error              { return nil }
func (f *fakeMemberUserRepo) MarkOnboarded(context.Context, uint64) error           { return nil }
func (f *fakeMemberUserRepo) UpdateAiChatEnabled(_ context.Context, userID uint64, enabled *bool) error {
	f.updated[userID] = enabled
	return nil
}

func ptrBool(b bool) *bool    { return &b }
func u64ptr(v uint64) *uint64 { return &v }

func Test_会社メンバー一覧ユースケース(t *testing.T) {
	repo := newFakeMemberUserRepo()
	repo.byCompany[10] = []domain.User{{ID: 1, CompanyID: u64ptr(10)}, {ID: 2, CompanyID: u64ptr(10)}}
	uc := usecase.NewListCompanyMembersUseCase(repo)

	t.Run("自社の従業員一覧を返す", func(t *testing.T) {
		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, CompanyID: u64ptr(10), Role: domain.RoleCompanyAdmin})
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})
	t.Run("会社未所属は空", func(t *testing.T) {
		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, Role: domain.RoleSuperAdmin})
		require.NoError(t, err)
		assert.Empty(t, members)
	})
}

func Test_メンバーAI利用可否更新ユースケース(t *testing.T) {
	repo := newFakeMemberUserRepo()
	repo.byID[1] = &domain.User{ID: 1, CompanyID: u64ptr(10), Role: domain.RoleTrainee}
	repo.byID[2] = &domain.User{ID: 2, CompanyID: u64ptr(20), Role: domain.RoleTrainee} // 別会社
	uc := usecase.NewUpdateMemberAiAccessUseCase(repo)
	actor := &domain.User{ID: 9, CompanyID: u64ptr(10), Role: domain.RoleCompanyAdmin}

	t.Run("自社の従業員の AI を個別 OFF にできる", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 1, ptrBool(false))
		require.NoError(t, err)
		require.NotNil(t, repo.updated[1])
		assert.False(t, *repo.updated[1])
	})
	t.Run("nil で会社設定に従う状態へ戻せる", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 1, nil)
		require.NoError(t, err)
		assert.Nil(t, repo.updated[1])
	})
	t.Run("別会社の従業員は更新できない(403相当)", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 2, ptrBool(true))
		assert.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	})
	t.Run("会社未所属の actor は更新できない", func(t *testing.T) {
		err := uc.Execute(context.Background(), &domain.User{ID: 9, Role: domain.RoleSuperAdmin}, 1, ptrBool(true))
		assert.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	})
}
