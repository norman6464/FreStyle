package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func u64ptr(v uint64) *uint64 { return &v }

func strptr(v string) *string { return &v }

const memberWorkspaceID = "0198a000-0000-7000-8000-000000000001"

func Test_会社メンバー一覧ユースケース(t *testing.T) {
	t.Run("自社の従業員一覧を返す", func(t *testing.T) {
		repo := &mockUserRepo{}
		repo.On("ListByWorkspaceID", mock.Anything, memberWorkspaceID).
			Return([]domain.User{{ID: 1, CompanyID: u64ptr(10)}, {ID: 2, CompanyID: u64ptr(10)}}, nil)
		uc := usecase.NewListCompanyMembersUseCase(repo)

		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, WorkspaceID: strptr(memberWorkspaceID), Role: domain.RoleCompanyAdmin})
		require.NoError(t, err)
		require.Len(t, members, 2)
		assert.Equal(t, uint64(1), members[0].ID)
		assert.Equal(t, uint64(2), members[1].ID)
	})

	t.Run("会社未所属は空（repository を呼ばない）", func(t *testing.T) {
		repo := &mockUserRepo{}
		uc := usecase.NewListCompanyMembersUseCase(repo)

		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, Role: domain.RoleSuperAdmin})
		require.NoError(t, err)
		assert.Empty(t, members)
		repo.AssertNotCalled(t, "ListByWorkspaceID", mock.Anything, mock.Anything)
	})

	t.Run("actor が nil なら空（repository を呼ばない）", func(t *testing.T) {
		repo := &mockUserRepo{}
		uc := usecase.NewListCompanyMembersUseCase(repo)

		members, err := uc.Execute(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, members)
		repo.AssertNotCalled(t, "ListByWorkspaceID", mock.Anything, mock.Anything)
	})

	t.Run("repository のエラーはそのまま伝える", func(t *testing.T) {
		wantErr := errors.New("db down")
		repo := &mockUserRepo{}
		repo.On("ListByWorkspaceID", mock.Anything, memberWorkspaceID).
			Return([]domain.User(nil), wantErr)
		uc := usecase.NewListCompanyMembersUseCase(repo)

		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, WorkspaceID: strptr(memberWorkspaceID), Role: domain.RoleCompanyAdmin})
		assert.ErrorIs(t, err, wantErr)
		assert.Nil(t, members)
	})
}
