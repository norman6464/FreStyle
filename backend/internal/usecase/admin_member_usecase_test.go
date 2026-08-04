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

// memberStore は fake が読み書きする in-memory の状態。
type memberStore struct {
	byCompany map[uint64][]domain.User
	byID      map[uint64]*domain.User
	updated   map[uint64]*bool
}

// memberUserRepo は UserRepository の fake を、このテストが使う 3 メソッドだけ差し込んで返す。
// 残り 9 メソッドは生成 fake がゼロ値を返すので、no-op を手で書く必要がない。
func memberUserRepo() (*repofakes.FakeUserRepository, *memberStore) {
	s := &memberStore{
		byCompany: map[uint64][]domain.User{},
		byID:      map[uint64]*domain.User{},
		updated:   map[uint64]*bool{},
	}
	repo := &repofakes.FakeUserRepository{
		FindByIDFunc: func(_ context.Context, id uint64) (*domain.User, error) {
			return s.byID[id], nil
		},
		ListByCompanyIDFunc: func(_ context.Context, companyID uint64) ([]domain.User, error) {
			return s.byCompany[companyID], nil
		},
		UpdateAiChatEnabledFunc: func(_ context.Context, userID uint64, enabled *bool) error {
			s.updated[userID] = enabled
			return nil
		},
	}
	return repo, s
}

func ptrBool(b bool) *bool    { return &b }
func u64ptr(v uint64) *uint64 { return &v }

func Test_会社メンバー一覧ユースケース(t *testing.T) {
	repo, store := memberUserRepo()
	store.byCompany[10] = []domain.User{{ID: 1, CompanyID: u64ptr(10)}, {ID: 2, CompanyID: u64ptr(10)}}
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
	repo, store := memberUserRepo()
	store.byID[1] = &domain.User{ID: 1, CompanyID: u64ptr(10), Role: domain.RoleTrainee}
	store.byID[2] = &domain.User{ID: 2, CompanyID: u64ptr(20), Role: domain.RoleTrainee} // 別会社
	uc := usecase.NewUpdateMemberAiAccessUseCase(repo)
	actor := &domain.User{ID: 9, CompanyID: u64ptr(10), Role: domain.RoleCompanyAdmin}

	t.Run("自社の従業員の AI を個別 OFF にできる", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 1, ptrBool(false))
		require.NoError(t, err)
		require.NotNil(t, store.updated[1])
		assert.False(t, *store.updated[1])
	})
	t.Run("nil で会社設定に従う状態へ戻せる", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 1, nil)
		require.NoError(t, err)
		assert.Nil(t, store.updated[1])
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
