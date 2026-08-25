package usecase_test

import (
	"context"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockWorkspaceProvisioner は repository.WorkspaceProvisioner のモック。
type mockWorkspaceProvisioner struct{ mock.Mock }

var _ repository.WorkspaceProvisioner = (*mockWorkspaceProvisioner)(nil)

func (m *mockWorkspaceProvisioner) ProvisionWorkspace(
	ctx context.Context, in repository.WorkspaceProvisionInput,
) (*domain.Workspace, error) {
	args := m.Called(ctx, in)
	ws, _ := args.Get(0).(*domain.Workspace)
	return ws, args.Error(1)
}

func Test_ワークスペース作成_入力の検証(t *testing.T) {
	// 弾かれた入力が provisioner まで届かないこと（届くと DB の CHECK 頼みになり 500 になる）。
	provisioner := &mockWorkspaceProvisioner{}
	uc := usecase.NewCreateWorkspaceUseCase(provisioner)
	ctx := context.Background()

	cases := []struct {
		name string
		in   usecase.CreateWorkspaceInput
	}{
		{name: "作成者が無い", in: usecase.CreateWorkspaceInput{Slug: "acme", Name: "Acme"}},
		{name: "slug が空", in: usecase.CreateWorkspaceInput{Name: "Acme", OwnerUserID: 1}},
		{name: "slug に大文字", in: usecase.CreateWorkspaceInput{Slug: "Acme", Name: "Acme", OwnerUserID: 1}},
		{name: "slug に記号", in: usecase.CreateWorkspaceInput{Slug: "acme inc", Name: "Acme", OwnerUserID: 1}},
		{name: "slug の先頭がハイフン", in: usecase.CreateWorkspaceInput{Slug: "-acme", Name: "Acme", OwnerUserID: 1}},
		{name: "slug の末尾がハイフン", in: usecase.CreateWorkspaceInput{Slug: "acme-", Name: "Acme", OwnerUserID: 1}},
		{
			name: "slug が列幅を超える",
			in: usecase.CreateWorkspaceInput{
				Slug: strings.Repeat("a", domain.WorkspaceSlugMaxLen+1), Name: "Acme", OwnerUserID: 1,
			},
		},
		{name: "名前が空", in: usecase.CreateWorkspaceInput{Slug: "acme", OwnerUserID: 1}},
		{
			name: "名前が列幅を超える",
			in: usecase.CreateWorkspaceInput{
				Slug: "acme", Name: strings.Repeat("あ", domain.WorkspaceNameMaxLen+1), OwnerUserID: 1,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Execute(ctx, tc.in)
			require.Error(t, err)
		})
	}
	provisioner.AssertNotCalled(t, "ProvisionWorkspace", mock.Anything, mock.Anything)
}

func Test_ワークスペース作成_名前は文字数で数える(t *testing.T) {
	provisioner := &mockWorkspaceProvisioner{}
	name := strings.Repeat("あ", domain.WorkspaceNameMaxLen)
	provisioner.On("ProvisionWorkspace", mock.Anything, repository.WorkspaceProvisionInput{
		Slug: "acme", Name: name, OwnerUserID: 1,
	}).Return(&domain.Workspace{ID: kbWS, Slug: "acme", Name: name}, nil)
	uc := usecase.NewCreateWorkspaceUseCase(provisioner)

	// varchar(200) は「文字数」の上限なので、バイト数で数えると日本語 200 文字を弾いてしまう。
	got, err := uc.Execute(context.Background(), usecase.CreateWorkspaceInput{
		Slug: "acme", Name: name, OwnerUserID: 1,
	})
	require.NoError(t, err)
	assert.Equal(t, "acme", got.Slug)
}

func Test_ワークスペース作成_作成者をそのまま渡す(t *testing.T) {
	// 作成者を admin のメンバーにするのは provisioner（1 トランザクション）の責務なので、
	// usecase の責任は「誰が作ったかを取り違えずに渡すこと」に尽きる。
	provisioner := &mockWorkspaceProvisioner{}
	provisioner.On("ProvisionWorkspace", mock.Anything, repository.WorkspaceProvisionInput{
		Slug: "new-team", Name: "新チーム", OwnerUserID: 42,
	}).Return(&domain.Workspace{ID: kbWS, Slug: "new-team", Name: "新チーム"}, nil)
	uc := usecase.NewCreateWorkspaceUseCase(provisioner)

	got, err := uc.Execute(context.Background(), usecase.CreateWorkspaceInput{
		Slug: "new-team", Name: "新チーム", OwnerUserID: 42,
	})
	require.NoError(t, err)
	assert.Equal(t, "new-team", got.Slug)
	provisioner.AssertExpectations(t)
}

func Test_ワークスペース作成_slug衝突はそのまま伝える(t *testing.T) {
	provisioner := &mockWorkspaceProvisioner{}
	provisioner.On("ProvisionWorkspace", mock.Anything, mock.Anything).
		Return(nil, repository.ErrWorkspaceSlugTaken)
	uc := usecase.NewCreateWorkspaceUseCase(provisioner)

	_, err := uc.Execute(context.Background(), usecase.CreateWorkspaceInput{
		Slug: "acme", Name: "Acme", OwnerUserID: 1,
	})
	assert.ErrorIs(t, err, repository.ErrWorkspaceSlugTaken)
}

func Test_スペース作成_入力の検証(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	uc := usecase.NewCreateSpaceUseCase(repo)
	ctx := context.Background()

	cases := []struct {
		name string
		in   usecase.CreateSpaceInput
	}{
		{name: "workspaceID が無い", in: usecase.CreateSpaceInput{Key: "eng", Name: "開発部"}},
		{name: "key が空", in: usecase.CreateSpaceInput{WorkspaceID: kbWS, Name: "開発部"}},
		{name: "key に大文字", in: usecase.CreateSpaceInput{WorkspaceID: kbWS, Key: "ENG", Name: "開発部"}},
		{
			name: "key が列幅を超える",
			in: usecase.CreateSpaceInput{
				WorkspaceID: kbWS, Key: strings.Repeat("a", domain.SpaceKeyMaxLen+1), Name: "開発部",
			},
		},
		{name: "名前が空", in: usecase.CreateSpaceInput{WorkspaceID: kbWS, Key: "eng"}},
		{
			name: "名前が列幅を超える",
			in: usecase.CreateSpaceInput{
				WorkspaceID: kbWS, Key: "eng", Name: strings.Repeat("あ", domain.SpaceNameMaxLen+1),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.Execute(ctx, tc.in)
			require.Error(t, err)
		})
	}
	repo.AssertNotCalled(t, "CreateSpace", mock.Anything, mock.Anything)
}

func Test_スペース作成_repositoryが確定させた行を返す(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("CreateSpace", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		space := args.Get(1).(*domain.Space)
		space.ID = kbSpace // ID の採番は repository の責務。
	}).Return(nil)
	uc := usecase.NewCreateSpaceUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CreateSpaceInput{
		WorkspaceID: kbWS, Key: "eng", Name: "開発部",
	})
	require.NoError(t, err)
	assert.Equal(t, kbSpace, got.ID)
	assert.Equal(t, kbWS, got.WorkspaceID)
	assert.Equal(t, "eng", got.Key)
}

func Test_スペース作成_key衝突はそのまま伝える(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	repo.On("CreateSpace", mock.Anything, mock.Anything).Return(repository.ErrSpaceKeyTaken)
	uc := usecase.NewCreateSpaceUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreateSpaceInput{
		WorkspaceID: kbWS, Key: "eng", Name: "開発部",
	})
	assert.ErrorIs(t, err, repository.ErrSpaceKeyTaken)
}
