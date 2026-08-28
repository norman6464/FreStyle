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

func Test_スペース改名_検証と伝播(t *testing.T) {
	t.Run("空の名前と文字数超過は repo に届く前に弾く", func(t *testing.T) {
		repo := &mockKnowledgeBaseRepo{}
		uc := usecase.NewRenameSpaceUseCase(repo)
		for _, name := range []string{"", strings.Repeat("あ", 201)} {
			_, err := uc.Execute(context.Background(), usecase.RenameSpaceInput{
				WorkspaceID: "ws-1", SpaceID: "sp-1", Name: name,
			})
			assert.ErrorIs(t, err, usecase.ErrInvalidName)
		}
		repo.AssertNotCalled(t, "UpdateSpaceName", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("200 文字ちょうどは通る（列幅は文字数）", func(t *testing.T) {
		repo := &mockKnowledgeBaseRepo{}
		name := strings.Repeat("あ", 200)
		repo.On("UpdateSpaceName", mock.Anything, "ws-1", "sp-1", name).Return(nil)
		repo.On("FindSpace", mock.Anything, "ws-1", "sp-1").
			Return(&domain.Space{ID: "sp-1", WorkspaceID: "ws-1", Key: "eng", Name: name}, nil)
		uc := usecase.NewRenameSpaceUseCase(repo)
		got, err := uc.Execute(context.Background(), usecase.RenameSpaceInput{
			WorkspaceID: "ws-1", SpaceID: "sp-1", Name: name,
		})
		require.NoError(t, err)
		assert.Equal(t, name, got.Name)
		// UpdateSpaceName を呼ばずに FindSpace の結果だけ返す回帰を捕まえる。
		repo.AssertExpectations(t)
	})

	t.Run("存在しないスペースは ErrSpaceNotFound をそのまま伝える", func(t *testing.T) {
		repo := &mockKnowledgeBaseRepo{}
		repo.On("UpdateSpaceName", mock.Anything, "ws-1", "sp-x", "新名").
			Return(repository.ErrSpaceNotFound)
		uc := usecase.NewRenameSpaceUseCase(repo)
		_, err := uc.Execute(context.Background(), usecase.RenameSpaceInput{
			WorkspaceID: "ws-1", SpaceID: "sp-x", Name: "新名",
		})
		assert.ErrorIs(t, err, repository.ErrSpaceNotFound)
		repo.AssertExpectations(t)
	})
}

func Test_ワークスペース作成_slugが空なら自動採番される(t *testing.T) {
	// URL に使う名前は利用者に決めさせない（ユーザー決定 2026-08-28）。
	// 空で渡すと形の正しい slug が生成され、そのまま provisioner に届くこと。
	provisioner := &mockWorkspaceProvisioner{}
	var got repository.WorkspaceProvisionInput
	provisioner.On("ProvisionWorkspace", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { got, _ = args.Get(1).(repository.WorkspaceProvisionInput) }).
		Return(&domain.Workspace{ID: "ws-1", Slug: "w-abc", Name: "Acme"}, nil)
	uc := usecase.NewCreateWorkspaceUseCase(provisioner)

	_, err := uc.Execute(context.Background(), usecase.CreateWorkspaceInput{Name: "Acme", OwnerUserID: 1})
	require.NoError(t, err)
	assert.True(t, domain.ValidWorkspaceSlug(got.Slug), "生成された slug %q が URL の規則を満たすこと", got.Slug)
	assert.True(t, strings.HasPrefix(got.Slug, "w-"))
}

func Test_スペース作成_keyが空なら自動採番される(t *testing.T) {
	repo := &mockKnowledgeBaseRepo{}
	var got *domain.Space
	repo.On("CreateSpace", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) { got, _ = args.Get(1).(*domain.Space) }).
		Return(nil)
	uc := usecase.NewCreateSpaceUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CreateSpaceInput{WorkspaceID: "ws-1", Name: "開発部"})
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.True(t, domain.ValidSpaceKey(got.Key), "生成された key %q が規則を満たすこと", got.Key)
	assert.True(t, strings.HasPrefix(got.Key, "s-"))
}
