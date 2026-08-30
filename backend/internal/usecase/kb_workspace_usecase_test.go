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

const kbSlug = "acme"

func Test_ワークスペース解決_所属していれば返る(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	perm := &mockKBPermissionRepo{}
	users := &mockUserRepo{}
	repo.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	perm.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(true, nil)
	uc := usecase.NewResolveWorkspaceUseCase(repo, perm, users)

	ws, err := uc.Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})

	require.NoError(t, err)
	assert.Equal(t, kbWS, ws.ID)
}

func Test_ワークスペース解決_未所属は存在しないのと同じ(t *testing.T) {
	member := &mockKnowledgeBaseRepo{}
	memberPerm := &mockKBPermissionRepo{}
	memberUsers := &mockUserRepo{}
	member.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	memberPerm.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(false, nil)
	// 非メンバーでも、会社のワークスペースなら自動で入る。ここでは「会社が違う」ので
	// 入れる先が無い（ErrWorkspaceNotFound）ことを表し、404 に倒れることを確かめる。
	memberUsers.On("FindByID", mock.Anything, uint64(1)).Return((*domain.User)(nil), nil)

	unknown := &mockKnowledgeBaseRepo{}
	unknown.On("FindWorkspaceBySlug", mock.Anything, "no-such").
		Return(nil, repository.ErrWorkspaceNotFound)

	_, foreignErr := usecase.NewResolveWorkspaceUseCase(member, memberPerm, memberUsers).
		Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})
	_, unknownErr := usecase.NewResolveWorkspaceUseCase(unknown, &mockKBPermissionRepo{}, &mockUserRepo{}).
		Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: "no-such", UserID: 1})

	require.ErrorIs(t, foreignErr, repository.ErrWorkspaceNotFound,
		"未所属と不在を撃ち分けない（slug の総当たりでテナントの実在を漏らさない）")
	require.ErrorIs(t, unknownErr, repository.ErrWorkspaceNotFound)
}

// 会社のワークスペースは、まだ principals の行が無くても開ける（その場で所属を用意する）。
// これが無いと、会社ごとのワークスペースを作成者以外の誰も開けない。
func Test_ワークスペース解決_会社のワークスペースなら所属を用意して通す(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	perms := &mockKBPermissionRepo{}
	users := &mockUserRepo{}
	repo.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	perms.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(false, nil)
	wsID := kbWS
	users.On("FindByID", mock.Anything, uint64(1)).Return(&domain.User{ID: 1, WorkspaceID: &wsID}, nil)
	perms.On("EnsureUserPrincipal", mock.Anything, kbWS, uint64(1)).
		Return(&domain.Principal{ID: "principal-1", WorkspaceID: kbWS}, nil)
	perms.On("GrantWorkspaceRoleIfAbsent", mock.Anything, kbWS, "principal-1", domain.GrantRoleEditor).
		Return(nil)

	ws, err := usecase.NewResolveWorkspaceUseCase(repo, perms, users).
		Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})

	require.NoError(t, err)
	assert.Equal(t, kbWS, ws.ID)
	perms.AssertExpectations(t)
}

func Test_ワークスペース解決_入力の検証(t *testing.T) {
	uc := usecase.NewResolveWorkspaceUseCase(&mockKnowledgeBaseRepo{}, &mockKBPermissionRepo{}, &mockUserRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.ResolveWorkspaceInput{Slug: kbSlug})
	require.Error(t, err, "userID 必須")

	_, err = uc.Execute(ctx, usecase.ResolveWorkspaceInput{UserID: 1})
	require.ErrorIs(t, err, repository.ErrWorkspaceNotFound, "空 slug は不在扱い")
}

func Test_ワークスペース解決_所属判定の失敗はそのまま返す(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	perm := &mockKBPermissionRepo{}
	boom := errors.New("db down")
	repo.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	perm.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(false, boom)
	uc := usecase.NewResolveWorkspaceUseCase(repo, perm, &mockUserRepo{})

	_, err := uc.Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})

	require.ErrorIs(t, err, boom, "DB 障害を「不在」に潰すと 500 が 404 に化ける")
}
