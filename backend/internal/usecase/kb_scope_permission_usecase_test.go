package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_スペース権限確認_必須項目の検証(t *testing.T) {
	uc := usecase.NewCheckSpacePermissionUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CheckSpacePermissionInput{SpaceID: kbSpace, UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CheckSpacePermissionInput{WorkspaceID: kbWS, UserID: 1})
	assert.ErrorIs(t, err, repository.ErrSpaceNotFound, "spaceID が空なら「無い」と同じ扱い")
	_, err = uc.Execute(ctx, usecase.CheckSpacePermissionInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.Error(t, err, "userID 必須")
}

// 集めた事実（役割の集合）を畳むのは domain.ResolveScopePermission であって、
// usecase は規則を持たない。強い方が採られることを usecase 経由で確かめる。
func Test_スペース権限確認_集めた役割を規則にかけて返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("SpacePermissionFactsForUser", mock.Anything, kbWS, kbSpace, uint64(1)).
		Return(&domain.ScopeFacts{Roles: []domain.GrantRole{
			domain.GrantRoleViewer, domain.GrantRoleEditor,
		}}, nil)
	uc := usecase.NewCheckSpacePermissionUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CheckSpacePermissionInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1,
	})
	require.NoError(t, err)
	assert.True(t, got.CanView)
	assert.True(t, got.CanEdit, "viewer と editor なら強い editor が効く")
	assert.False(t, got.CanManage)
}

func Test_スペース権限確認_役割が1つも無ければ何もできない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("SpacePermissionFactsForUser", mock.Anything, kbWS, kbSpace, uint64(1)).
		Return(&domain.ScopeFacts{}, nil)
	uc := usecase.NewCheckSpacePermissionUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CheckSpacePermissionInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.ScopePermission{}, *got, "fail-closed（見えないものは作れない）")
}

func Test_スペース権限確認_スペースが無ければそのまま伝える(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("SpacePermissionFactsForUser", mock.Anything, kbWS, kbSpace, uint64(1)).
		Return(nil, repository.ErrSpaceNotFound)
	uc := usecase.NewCheckSpacePermissionUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CheckSpacePermissionInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1,
	})
	assert.ErrorIs(t, err, repository.ErrSpaceNotFound)
}

func Test_ワークスペース権限確認_必須項目の検証(t *testing.T) {
	uc := usecase.NewCheckWorkspacePermissionUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CheckWorkspacePermissionInput{UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CheckWorkspacePermissionInput{WorkspaceID: kbWS})
	require.Error(t, err, "userID 必須")
}

func Test_ワークスペース権限確認_集めた役割を規則にかけて返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("WorkspacePermissionFactsForUser", mock.Anything, kbWS, uint64(1)).
		Return(&domain.ScopeFacts{Roles: []domain.GrantRole{domain.GrantRoleAdmin}}, nil)
	uc := usecase.NewCheckWorkspacePermissionUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CheckWorkspacePermissionInput{
		WorkspaceID: kbWS, UserID: 1,
	})
	require.NoError(t, err)
	assert.True(t, got.CanManage)
}

func Test_所属ワークスペース一覧_必須項目の検証(t *testing.T) {
	uc := usecase.NewListMemberWorkspacesUseCase(&mockKBPermissionRepo{})

	_, err := uc.Execute(context.Background(), usecase.ListMemberWorkspacesInput{})
	require.Error(t, err, "userID 必須")
}

func Test_所属ワークスペース一覧_repositoryの結果をそのまま返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("ListMemberWorkspaces", mock.Anything, uint64(7)).
		Return([]domain.MemberWorkspace{{Workspace: domain.Workspace{ID: kbWS, Slug: "acme"}, CanManage: true}}, nil)
	uc := usecase.NewListMemberWorkspacesUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.ListMemberWorkspacesInput{UserID: 7})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "acme", got[0].Slug)
	assert.True(t, got[0].CanManage)
}

func Test_所属ワークスペース一覧_失敗はそのまま伝える(t *testing.T) {
	wantErr := errors.New("db down")
	repo := &mockKBPermissionRepo{}
	repo.On("ListMemberWorkspaces", mock.Anything, uint64(7)).
		Return([]domain.MemberWorkspace(nil), wantErr)
	uc := usecase.NewListMemberWorkspacesUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.ListMemberWorkspacesInput{UserID: 7})
	assert.ErrorIs(t, err, wantErr)
}
