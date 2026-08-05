package usecase_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ptrBool(b bool) *bool    { return &b }
func u64ptr(v uint64) *uint64 { return &v }

func Test_会社メンバー一覧ユースケース(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("ListByCompanyID", mock.Anything, uint64(10)).
		Return([]domain.User{{ID: 1, CompanyID: u64ptr(10)}, {ID: 2, CompanyID: u64ptr(10)}}, nil).
		Maybe()
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
	repo := &mockUserRepo{}
	repo.On("FindByID", mock.Anything, uint64(1)).
		Return(&domain.User{ID: 1, CompanyID: u64ptr(10), Role: domain.RoleTrainee}, nil).Maybe()
	repo.On("FindByID", mock.Anything, uint64(2)).
		Return(&domain.User{ID: 2, CompanyID: u64ptr(20), Role: domain.RoleTrainee}, nil).Maybe() // 別会社

	// 渡された *bool は Run で控えて中身まで検証する(nil で「会社設定に従う」に戻せるため)。
	var updateCalled bool
	var lastEnabled *bool
	repo.On("UpdateAiChatEnabled", mock.Anything, uint64(1), mock.Anything).
		Run(func(args mock.Arguments) {
			updateCalled = true
			lastEnabled, _ = args.Get(2).(*bool)
		}).Return(nil).Maybe()

	uc := usecase.NewUpdateMemberAiAccessUseCase(repo)
	actor := &domain.User{ID: 9, CompanyID: u64ptr(10), Role: domain.RoleCompanyAdmin}

	t.Run("自社の従業員の AI を個別 OFF にできる", func(t *testing.T) {
		updateCalled, lastEnabled = false, nil
		err := uc.Execute(context.Background(), actor, 1, ptrBool(false))
		require.NoError(t, err)
		require.True(t, updateCalled, "UpdateAiChatEnabled が呼ばれていること")
		require.NotNil(t, lastEnabled)
		assert.False(t, *lastEnabled)
	})
	t.Run("nil で会社設定に従う状態へ戻せる", func(t *testing.T) {
		updateCalled, lastEnabled = false, nil
		err := uc.Execute(context.Background(), actor, 1, nil)
		require.NoError(t, err)
		require.True(t, updateCalled, "UpdateAiChatEnabled が呼ばれていること")
		assert.Nil(t, lastEnabled)
	})
	t.Run("別会社の従業員は更新できない(403相当)", func(t *testing.T) {
		err := uc.Execute(context.Background(), actor, 2, ptrBool(true))
		assert.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
		repo.AssertNotCalled(t, "UpdateAiChatEnabled", mock.Anything, uint64(2), mock.Anything)
	})
	t.Run("会社未所属の actor は更新できない", func(t *testing.T) {
		err := uc.Execute(context.Background(), &domain.User{ID: 9, Role: domain.RoleSuperAdmin}, 1, ptrBool(true))
		assert.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany)
	})
	t.Run("自分自身のAI利用可否も更新できる(自己編集の禁止は無い)", func(t *testing.T) {
		updateCalled, lastEnabled = false, nil
		selfActor := &domain.User{ID: 1, CompanyID: u64ptr(10), Role: domain.RoleCompanyAdmin}
		err := uc.Execute(context.Background(), selfActor, 1, ptrBool(false))
		require.NoError(t, err, "UpdateMemberAiAccessUseCase には自己編集を禁止するチェックが無い")
		require.True(t, updateCalled)
		require.NotNil(t, lastEnabled)
		assert.False(t, *lastEnabled)
	})
	t.Run("super_admin(会社未所属)はこのusecaseを実行できない", func(t *testing.T) {
		superActor := &domain.User{ID: 9, Role: domain.RoleSuperAdmin, CompanyID: nil}
		err := uc.Execute(context.Background(), superActor, 1, ptrBool(true))
		assert.ErrorIs(t, err, usecase.ErrMemberNotInActorCompany,
			"super_admin は CompanyID=nil のため actor.CompanyID==nil チェックで弾かれる(現状の既知の制約)")
	})
}
