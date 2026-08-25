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
	repo.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	perm.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(true, nil)
	uc := usecase.NewResolveWorkspaceUseCase(repo, perm)

	ws, err := uc.Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})

	require.NoError(t, err)
	assert.Equal(t, kbWS, ws.ID)
}

func Test_ワークスペース解決_未所属は存在しないのと同じ(t *testing.T) {
	member := &mockKnowledgeBaseRepo{}
	memberPerm := &mockKBPermissionRepo{}
	member.On("FindWorkspaceBySlug", mock.Anything, kbSlug).
		Return(&domain.Workspace{ID: kbWS, Slug: kbSlug}, nil)
	memberPerm.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(false, nil)

	unknown := &mockKnowledgeBaseRepo{}
	unknown.On("FindWorkspaceBySlug", mock.Anything, "no-such").
		Return(nil, repository.ErrWorkspaceNotFound)

	_, foreignErr := usecase.NewResolveWorkspaceUseCase(member, memberPerm).
		Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})
	_, unknownErr := usecase.NewResolveWorkspaceUseCase(unknown, &mockKBPermissionRepo{}).
		Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: "no-such", UserID: 1})

	require.ErrorIs(t, foreignErr, repository.ErrWorkspaceNotFound,
		"未所属と不在を撃ち分けない（slug の総当たりでテナントの実在を漏らさない）")
	require.ErrorIs(t, unknownErr, repository.ErrWorkspaceNotFound)
}

func Test_ワークスペース解決_入力の検証(t *testing.T) {
	uc := usecase.NewResolveWorkspaceUseCase(&mockKnowledgeBaseRepo{}, &mockKBPermissionRepo{})
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
	uc := usecase.NewResolveWorkspaceUseCase(repo, perm)

	_, err := uc.Execute(context.Background(), usecase.ResolveWorkspaceInput{Slug: kbSlug, UserID: 1})

	require.ErrorIs(t, err, boom, "DB 障害を「不在」に潰すと 500 が 404 に化ける")
}
