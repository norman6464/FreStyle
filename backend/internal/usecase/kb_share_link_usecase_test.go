package usecase_test

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

const kbShareLink = "0198a000-0000-7000-8000-00000000000c"

func kbShareLinkFor(token string) *domain.ShareLink {
	sum := sha256.Sum256([]byte(token))
	return &domain.ShareLink{
		ID: kbShareLink, WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
		Capability: domain.CapabilityView, TokenHash: sum[:], CreatedByUserID: 1,
	}
}

func Test_共有リンク発行_平文トークンは一度だけ返りDBにはハッシュだけ渡す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	var got repository.ShareLinkWrite
	repo.On("CreateShareLink", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { got = args.Get(1).(repository.ShareLinkWrite) }).
		Return(&domain.ShareLink{ID: kbShareLink, WorkspaceID: kbWS, PageID: kbPage}, nil)
	uc := usecase.NewIssueShareLinkUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.IssueShareLinkInput{
		WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityView, CreatedByUserID: 1,
	})
	require.NoError(t, err)
	require.NotEmpty(t, out.Token, "平文トークンはこのときだけ返る")
	assert.Len(t, got.TokenHash, sha256.Size, "DB へ渡すのは SHA-256 の 32 バイト")
	assert.NotContains(t, string(got.TokenHash), out.Token, "平文は渡さない")

	sum := sha256.Sum256([]byte(out.Token))
	assert.Equal(t, sum[:], got.TokenHash, "保存するのは返したトークンのハッシュ")
	assert.Nil(t, got.PasswordHash, "パスワード無しなら nil")
}

func Test_共有リンク発行_パスワードはbcryptで持つ(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	var got repository.ShareLinkWrite
	repo.On("CreateShareLink", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { got = args.Get(1).(repository.ShareLinkWrite) }).
		Return(&domain.ShareLink{ID: kbShareLink}, nil)
	uc := usecase.NewIssueShareLinkUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.IssueShareLinkInput{
		WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityEdit,
		Password: "s3cret", CreatedByUserID: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, got.PasswordHash)
	assert.NotEqual(t, "s3cret", *got.PasswordHash, "平文は保存しない")
	require.NoError(t, bcrypt.CompareHashAndPassword([]byte(*got.PasswordHash), []byte("s3cret")))
}

func Test_共有リンク発行_入力の検証(t *testing.T) {
	uc := usecase.NewIssueShareLinkUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()
	past := time.Now().Add(-time.Hour)

	_, err := uc.Execute(ctx, usecase.IssueShareLinkInput{PageID: kbPage, Capability: domain.CapabilityView, CreatedByUserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.IssueShareLinkInput{WorkspaceID: kbWS, Capability: domain.CapabilityView, CreatedByUserID: 1})
	require.Error(t, err, "pageID 必須")
	_, err = uc.Execute(ctx, usecase.IssueShareLinkInput{WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityView})
	require.Error(t, err, "createdByUserID 必須")
	_, err = uc.Execute(ctx, usecase.IssueShareLinkInput{
		WorkspaceID: kbWS, PageID: kbPage, Capability: domain.Capability("delete"), CreatedByUserID: 1,
	})
	require.ErrorIs(t, err, usecase.ErrInvalidCapability)
	_, err = uc.Execute(ctx, usecase.IssueShareLinkInput{
		WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityView,
		ExpiresAt: &past, CreatedByUserID: 1,
	})
	require.Error(t, err, "過去の期限は受け付けない")
}

func Test_共有リンク検証_トークンが違えば見つからない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindShareLinkByTokenHash", mock.Anything, mock.Anything).Return(nil, repository.ErrShareLinkNotFound)
	uc := usecase.NewVerifyShareLinkUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.VerifyShareLinkInput{Token: "wrong"})
	require.ErrorIs(t, err, repository.ErrShareLinkNotFound)

	_, err = uc.Execute(context.Background(), usecase.VerifyShareLinkInput{})
	require.ErrorIs(t, err, repository.ErrShareLinkNotFound, "空のトークンは DB を引かずに not found")
}

func Test_共有リンク検証_期限切れと失効を区別する(t *testing.T) {
	past := time.Now().Add(-time.Hour)

	expired := kbShareLinkFor("tok")
	expired.ExpiresAt = &past
	repoExpired := &mockKBPermissionRepo{}
	repoExpired.On("FindShareLinkByTokenHash", mock.Anything, mock.Anything).Return(expired, nil)
	_, err := usecase.NewVerifyShareLinkUseCase(repoExpired).
		Execute(context.Background(), usecase.VerifyShareLinkInput{Token: "tok"})
	require.ErrorIs(t, err, usecase.ErrShareLinkExpired)

	revoked := kbShareLinkFor("tok")
	revoked.RevokedAt = &past
	repoRevoked := &mockKBPermissionRepo{}
	repoRevoked.On("FindShareLinkByTokenHash", mock.Anything, mock.Anything).Return(revoked, nil)
	_, err = usecase.NewVerifyShareLinkUseCase(repoRevoked).
		Execute(context.Background(), usecase.VerifyShareLinkInput{Token: "tok"})
	require.ErrorIs(t, err, usecase.ErrShareLinkRevoked)
}

func Test_共有リンク検証_パスワードの照合(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("s3cret"), bcrypt.MinCost)
	require.NoError(t, err)
	s := string(hash)

	link := kbShareLinkFor("tok")
	link.PasswordHash = &s
	repo := &mockKBPermissionRepo{}
	repo.On("FindShareLinkByTokenHash", mock.Anything, mock.Anything).Return(link, nil)
	uc := usecase.NewVerifyShareLinkUseCase(repo)
	ctx := context.Background()

	_, err = uc.Execute(ctx, usecase.VerifyShareLinkInput{Token: "tok"})
	require.ErrorIs(t, err, usecase.ErrShareLinkPasswordRequired)

	_, err = uc.Execute(ctx, usecase.VerifyShareLinkInput{Token: "tok", Password: "wrong"})
	require.ErrorIs(t, err, usecase.ErrShareLinkPasswordMismatch)

	got, err := uc.Execute(ctx, usecase.VerifyShareLinkInput{Token: "tok", Password: "s3cret"})
	require.NoError(t, err)
	assert.Equal(t, kbShareLink, got.ID)
}

func Test_共有リンク失効_必須項目の検証(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("RevokeShareLink", mock.Anything, kbWS, kbShareLink).Return(nil)
	uc := usecase.NewRevokeShareLinkUseCase(repo)
	ctx := context.Background()

	require.Error(t, uc.Execute(ctx, usecase.RevokeShareLinkInput{ShareLinkID: kbShareLink}))
	require.Error(t, uc.Execute(ctx, usecase.RevokeShareLinkInput{WorkspaceID: kbWS}))
	require.NoError(t, uc.Execute(ctx, usecase.RevokeShareLinkInput{WorkspaceID: kbWS, ShareLinkID: kbShareLink}))
}

func Test_共有リンク権限_対象ページの外は拒否(t *testing.T) {
	pages := &mockKnowledgeBaseRepo{}
	pages.On("HasDescendant", mock.Anything, kbWS, kbPage, "other").Return(false, nil)
	perms := &mockKBPermissionRepo{}
	uc := usecase.NewCheckShareLinkPermissionUseCase(perms, pages)

	_, err := uc.Execute(context.Background(), usecase.CheckShareLinkPermissionInput{
		Link: kbShareLinkFor("tok"), PageID: "other",
	})
	require.ErrorIs(t, err, usecase.ErrShareLinkPageOutOfScope)
	perms.AssertNotCalled(t, "PagePermissionFactsForPrincipal", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_共有リンク権限_リンクの既定を出発点に例外を適用する(t *testing.T) {
	link := kbShareLinkFor("tok")
	link.Capability = domain.CapabilityEdit
	child := "0198a000-0000-7000-8000-00000000000d"

	pages := &mockKnowledgeBaseRepo{}
	pages.On("HasDescendant", mock.Anything, kbWS, kbPage, child).Return(true, nil)
	perms := &mockKBPermissionRepo{}
	perms.On("PagePermissionFactsForPrincipal", mock.Anything, kbWS, child, kbPrincipal).
		Return(&domain.PagePermissionFacts{}, nil)
	uc := usecase.NewCheckShareLinkPermissionUseCase(perms, pages)

	got, err := uc.Execute(context.Background(), usecase.CheckShareLinkPermissionInput{Link: link, PageID: child})
	require.NoError(t, err)
	assert.True(t, got.CanView)
	assert.True(t, got.CanEdit, "編集可のリンクなので子ページも編集できる")
}

func Test_共有リンク権限_子ページのdenyで隠せる(t *testing.T) {
	child := "0198a000-0000-7000-8000-00000000000d"
	pages := &mockKnowledgeBaseRepo{}
	pages.On("HasDescendant", mock.Anything, kbWS, kbPage, child).Return(true, nil)
	perms := &mockKBPermissionRepo{}
	perms.On("PagePermissionFactsForPrincipal", mock.Anything, kbWS, child, kbPrincipal).
		Return(&domain.PagePermissionFacts{View: &domain.RestrictionFacts{DeniedAnywhere: true}}, nil)
	uc := usecase.NewCheckShareLinkPermissionUseCase(perms, pages)

	got, err := uc.Execute(context.Background(), usecase.CheckShareLinkPermissionInput{
		Link: kbShareLinkFor("tok"), PageID: child,
	})
	require.NoError(t, err)
	assert.False(t, got.CanView, "公開リンクの主体を deny した子ページは開けない")
}

func Test_共有リンク権限_必須項目の検証(t *testing.T) {
	uc := usecase.NewCheckShareLinkPermissionUseCase(&mockKBPermissionRepo{}, &mockKnowledgeBaseRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CheckShareLinkPermissionInput{PageID: kbPage})
	require.Error(t, err, "link 必須")
	_, err = uc.Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: kbShareLinkFor("tok")})
	require.Error(t, err, "pageID 必須")
}

func Test_共有リンク一覧_失効済みも含めて返す(t *testing.T) {
	revoked := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	repo := &mockKBPermissionRepo{}
	repo.On("ListPageShareLinks", mock.Anything, kbWS, kbPage).
		Return([]domain.ShareLink{
			{ID: "l1", WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityView},
			{ID: "l2", WorkspaceID: kbWS, PageID: kbPage, Capability: domain.CapabilityView, RevokedAt: &revoked},
		}, nil)
	uc := usecase.NewListPageShareLinksUseCase(repo)

	links, err := uc.Execute(context.Background(), usecase.ListPageShareLinksInput{
		WorkspaceID: kbWS, PageID: kbPage,
	})
	require.NoError(t, err)
	require.Len(t, links, 2, "止めたことの確認ができるよう失効済みも返す")
}

func Test_共有リンク一覧_必須項目の検証(t *testing.T) {
	uc := usecase.NewListPageShareLinksUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.ListPageShareLinksInput{PageID: kbPage})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.ListPageShareLinksInput{WorkspaceID: kbWS})
	require.Error(t, err, "pageID 必須")
}
