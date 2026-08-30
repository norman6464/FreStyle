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

func u64ptr(v uint64) *uint64 { return &v }

func strptr(v string) *string { return &v }

const memberWorkspaceID = "0198a000-0000-7000-8000-000000000001"

func Test_会社メンバー一覧ユースケース(t *testing.T) {
	repo := &mockUserRepo{}
	repo.On("ListByWorkspaceID", mock.Anything, memberWorkspaceID).
		Return([]domain.User{{ID: 1, CompanyID: u64ptr(10)}, {ID: 2, CompanyID: u64ptr(10)}}, nil).
		Maybe()
	uc := usecase.NewListCompanyMembersUseCase(repo)

	t.Run("自社の従業員一覧を返す", func(t *testing.T) {
		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, WorkspaceID: strptr(memberWorkspaceID), Role: domain.RoleCompanyAdmin})
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})
	t.Run("会社未所属は空", func(t *testing.T) {
		members, err := uc.Execute(context.Background(), &domain.User{ID: 9, Role: domain.RoleSuperAdmin})
		require.NoError(t, err)
		assert.Empty(t, members)
	})
}
