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

const (
	kbPrincipal = "0198a000-0000-7000-8000-00000000000a"
	kbGroup     = "0198a000-0000-7000-8000-00000000000b"
)

func kbGrantRole(r domain.GrantRole) *domain.GrantRole { return &r }

func Test_ページ権限確認_必須項目の検証(t *testing.T) {
	uc := usecase.NewCheckPagePermissionUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CheckPagePermissionInput{PageID: kbPage, UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CheckPagePermissionInput{WorkspaceID: kbWS, UserID: 1})
	require.Error(t, err, "pageID 必須")
	_, err = uc.Execute(ctx, usecase.CheckPagePermissionInput{WorkspaceID: kbWS, PageID: kbPage})
	require.Error(t, err, "userID 必須")
}

func Test_ページ権限確認_集めた事実を規則にかけて返す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("PagePermissionFactsForUser", mock.Anything, kbWS, kbPage, uint64(1)).
		Return(&domain.PagePermissionFacts{
			Member: true,
			Role:   kbGrantRole(domain.GrantRoleEditor),
			View:   &domain.RestrictionFacts{DeniedAnywhere: true},
		}, nil)
	uc := usecase.NewCheckPagePermissionUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CheckPagePermissionInput{
		WorkspaceID: kbWS, PageID: kbPage, UserID: 1,
	})
	require.NoError(t, err)
	assert.False(t, got.CanView, "祖先の deny が効く")
	assert.False(t, got.CanEdit, "閲覧できないので編集もできない")
}

func Test_ページ権限確認_ページが無ければそのまま伝える(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("PagePermissionFactsForUser", mock.Anything, kbWS, kbPage, uint64(1)).
		Return(nil, repository.ErrPageNotFound)
	uc := usecase.NewCheckPagePermissionUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.CheckPagePermissionInput{
		WorkspaceID: kbWS, PageID: kbPage, UserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrPageNotFound)
}

func Test_ワークスペース所属判定(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("IsWorkspaceMember", mock.Anything, kbWS, uint64(1)).Return(true, nil)
	uc := usecase.NewIsWorkspaceMemberUseCase(repo)

	ok, err := uc.Execute(context.Background(), usecase.IsWorkspaceMemberInput{WorkspaceID: kbWS, UserID: 1})
	require.NoError(t, err)
	assert.True(t, ok)

	_, err = uc.Execute(context.Background(), usecase.IsWorkspaceMemberInput{UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), usecase.IsWorkspaceMemberInput{WorkspaceID: kbWS})
	require.Error(t, err, "userID 必須")
}

func Test_閲覧可能ページ一覧_見えないページを落とす(t *testing.T) {
	visible := domain.Page{ID: "p1", WorkspaceID: kbWS, SpaceID: kbSpace, Title: "見える"}
	hidden := domain.Page{ID: "p2", WorkspaceID: kbWS, SpaceID: kbSpace, Title: "隠れる"}
	repo := &mockKBPermissionRepo{}
	repo.On("ListSpacePageViewFacts", mock.Anything, kbWS, kbSpace, uint64(1)).
		Return([]repository.PageWithViewFacts{
			{Page: visible, Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)}},
			{Page: hidden, Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{HasAllowList: true},
			}},
		}, nil)
	uc := usecase.NewListViewablePagesUseCase(repo)

	pages, err := uc.Execute(context.Background(), usecase.ListViewablePagesInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1,
	})
	require.NoError(t, err)
	require.Len(t, pages, 1)
	assert.Equal(t, "見える", pages[0].Title)
}

func Test_閲覧可能ページ一覧_必須項目の検証(t *testing.T) {
	uc := usecase.NewListViewablePagesUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.ListViewablePagesInput{SpaceID: kbSpace, UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.ListViewablePagesInput{WorkspaceID: kbWS, UserID: 1})
	require.Error(t, err, "spaceID 必須")
	_, err = uc.Execute(ctx, usecase.ListViewablePagesInput{WorkspaceID: kbWS, SpaceID: kbSpace})
	require.Error(t, err, "userID 必須")
}

func Test_サブツリー編集可否_必須項目の検証(t *testing.T) {
	uc := usecase.NewCanEditPageSubtreeUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CanEditPageSubtreeInput{PageID: kbPage, UserID: 1})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.CanEditPageSubtreeInput{WorkspaceID: kbWS, UserID: 1})
	require.Error(t, err, "pageID 必須")
	_, err = uc.Execute(ctx, usecase.CanEditPageSubtreeInput{WorkspaceID: kbWS, PageID: kbPage})
	require.Error(t, err, "userID 必須")
}

func Test_サブツリー編集可否_1枚でも編集できなければ不可(t *testing.T) {
	editable := domain.PagePermissionFacts{Member: true, Role: kbGrantRole(domain.GrantRoleEditor)}
	cases := map[string]struct {
		rows []repository.PageWithPermissionFacts
		want bool
	}{
		"全部編集できる": {
			rows: []repository.PageWithPermissionFacts{
				{PageID: kbPage, Facts: editable},
				{PageID: kbPage + "1", Facts: editable},
			},
			want: true,
		},
		"子孫の編集が外されている": {
			rows: []repository.PageWithPermissionFacts{
				{PageID: kbPage, Facts: editable},
				{PageID: kbPage + "1", Facts: domain.PagePermissionFacts{
					Member: true, Role: kbGrantRole(domain.GrantRoleEditor),
					Edit: &domain.RestrictionFacts{DeniedAnywhere: true},
				}},
			},
			want: false,
		},
		"子孫が閲覧すらできない": {
			rows: []repository.PageWithPermissionFacts{
				{PageID: kbPage, Facts: editable},
				{PageID: kbPage + "1", Facts: domain.PagePermissionFacts{
					Member: true, Role: kbGrantRole(domain.GrantRoleEditor),
					View: &domain.RestrictionFacts{DeniedAnywhere: true},
				}},
			},
			want: false,
		},
		"1 行も返らない（ページが無い）": {
			rows: []repository.PageWithPermissionFacts{},
			want: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			repo := &mockKBPermissionRepo{}
			repo.On("ListSubtreePagePermissionFacts", mock.Anything, kbWS, kbPage, uint64(1)).
				Return(tc.rows, nil)
			uc := usecase.NewCanEditPageSubtreeUseCase(repo)

			got, err := uc.Execute(context.Background(), usecase.CanEditPageSubtreeInput{
				WorkspaceID: kbWS, PageID: kbPage, UserID: 1,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func Test_サブツリー編集可否_事実の収集が失敗したら伝える(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("ListSubtreePagePermissionFacts", mock.Anything, kbWS, kbPage, uint64(1)).
		Return(nil, repository.ErrPageNotFound)
	uc := usecase.NewCanEditPageSubtreeUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.CanEditPageSubtreeInput{
		WorkspaceID: kbWS, PageID: kbPage, UserID: 1,
	})
	require.ErrorIs(t, err, repository.ErrPageNotFound)
	assert.False(t, got, "確認できないなら許可に倒さない")
}

func Test_メンバー追加_主体を作る(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("EnsureUserPrincipal", mock.Anything, kbWS, uint64(7)).
		Return(&domain.Principal{ID: kbPrincipal, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	uc := usecase.NewAddWorkspaceMemberUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.AddWorkspaceMemberInput{WorkspaceID: kbWS, UserID: 7})
	require.NoError(t, err)
	assert.Equal(t, domain.PrincipalKindUser, got.Kind)

	_, err = uc.Execute(context.Background(), usecase.AddWorkspaceMemberInput{UserID: 7})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(context.Background(), usecase.AddWorkspaceMemberInput{WorkspaceID: kbWS})
	require.Error(t, err, "userID 必須")
}

func Test_メンバー削除_非メンバーなら何もしない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(7)).Return(nil, repository.ErrPrincipalNotFound)
	uc := usecase.NewRemoveWorkspaceMemberUseCase(repo)

	require.NoError(t, uc.Execute(context.Background(), usecase.RemoveWorkspaceMemberInput{WorkspaceID: kbWS, UserID: 7}))
	repo.AssertNotCalled(t, "DeletePrincipal", mock.Anything, mock.Anything, mock.Anything)
}

func Test_メンバー削除_主体を消す(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(7)).
		Return(&domain.Principal{ID: kbPrincipal, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	repo.On("DeletePrincipal", mock.Anything, kbWS, kbPrincipal).Return(nil)
	uc := usecase.NewRemoveWorkspaceMemberUseCase(repo)

	require.NoError(t, uc.Execute(context.Background(), usecase.RemoveWorkspaceMemberInput{WorkspaceID: kbWS, UserID: 7}))
	repo.AssertCalled(t, "DeletePrincipal", mock.Anything, kbWS, kbPrincipal)
}

func Test_グループ作成_名前の検証(t *testing.T) {
	uc := usecase.NewCreatePrincipalGroupUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.CreatePrincipalGroupInput{WorkspaceID: kbWS})
	require.Error(t, err, "name 必須")
	_, err = uc.Execute(ctx, usecase.CreatePrincipalGroupInput{WorkspaceID: kbWS, Name: strings.Repeat("あ", 201)})
	require.Error(t, err, "name は 200 文字まで")
	_, err = uc.Execute(ctx, usecase.CreatePrincipalGroupInput{Name: "開発"})
	require.Error(t, err, "workspaceID 必須")
}

func Test_グループ所属追加_グループでない主体は拒否(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbGroup).
		Return(&domain.Principal{ID: kbGroup, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	uc := usecase.NewAddGroupMemberUseCase(repo)

	err := uc.Execute(context.Background(), usecase.AddGroupMemberInput{
		WorkspaceID: kbWS, GroupPrincipalID: kbGroup, MemberUserID: 7,
	})
	require.ErrorIs(t, err, usecase.ErrPrincipalKindMismatch)
	repo.AssertNotCalled(t, "AddGroupMember", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_グループ所属追加_非メンバーは加えられない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbGroup).
		Return(&domain.Principal{ID: kbGroup, WorkspaceID: kbWS, Kind: domain.PrincipalKindGroup, Name: "開発"}, nil)
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(7)).Return(nil, repository.ErrPrincipalNotFound)
	uc := usecase.NewAddGroupMemberUseCase(repo)

	err := uc.Execute(context.Background(), usecase.AddGroupMemberInput{
		WorkspaceID: kbWS, GroupPrincipalID: kbGroup, MemberUserID: 7,
	})
	require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
}

func Test_グループ所属削除_非メンバーなら何もしない(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindUserPrincipal", mock.Anything, kbWS, uint64(7)).Return(nil, repository.ErrPrincipalNotFound)
	uc := usecase.NewRemoveGroupMemberUseCase(repo)

	require.NoError(t, uc.Execute(context.Background(), usecase.RemoveGroupMemberInput{
		WorkspaceID: kbWS, GroupPrincipalID: kbGroup, MemberUserID: 7,
	}))
	repo.AssertNotCalled(t, "RemoveGroupMember", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_スペース全員の主体_必須項目の検証(t *testing.T) {
	uc := usecase.NewEnsureSpaceEveryonePrincipalUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{SpaceID: kbSpace})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{WorkspaceID: kbWS})
	require.Error(t, err, "spaceID 必須")
}

func Test_権限付与_役割の検証(t *testing.T) {
	ctx := context.Background()

	wsUC := usecase.NewGrantWorkspaceRoleUseCase(&mockKBPermissionRepo{})
	_, err := wsUC.Execute(ctx, usecase.GrantWorkspaceRoleInput{
		WorkspaceID: kbWS, PrincipalID: kbPrincipal, Role: domain.GrantRole("owner"),
	})
	require.ErrorIs(t, err, usecase.ErrInvalidGrantRole)

	spUC := usecase.NewGrantSpaceRoleUseCase(&mockKBPermissionRepo{})
	_, err = spUC.Execute(ctx, usecase.GrantSpaceRoleInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, PrincipalID: kbPrincipal, Role: domain.GrantRole(""),
	})
	require.ErrorIs(t, err, usecase.ErrInvalidGrantRole)
}

func Test_権限付与_別ワークスペースの主体は拒否(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbPrincipal).Return(nil, repository.ErrPrincipalNotFound)
	uc := usecase.NewGrantSpaceRoleUseCase(repo)

	_, err := uc.Execute(context.Background(), usecase.GrantSpaceRoleInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, PrincipalID: kbPrincipal, Role: domain.GrantRoleEditor,
	})
	require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
	repo.AssertNotCalled(t, "UpsertSpaceGrant", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_権限付与_ワークスペースとスペースの両方に張れる(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbPrincipal).
		Return(&domain.Principal{ID: kbPrincipal, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	repo.On("UpsertWorkspaceGrant", mock.Anything, kbWS, kbPrincipal, domain.GrantRoleAdmin).
		Return(&domain.WorkspaceGrant{WorkspaceID: kbWS, PrincipalID: kbPrincipal, Role: domain.GrantRoleAdmin}, nil)
	repo.On("UpsertSpaceGrant", mock.Anything, kbWS, kbSpace, kbPrincipal, domain.GrantRoleViewer).
		Return(&domain.SpaceGrant{WorkspaceID: kbWS, SpaceID: kbSpace, PrincipalID: kbPrincipal, Role: domain.GrantRoleViewer}, nil)

	wsGrant, err := usecase.NewGrantWorkspaceRoleUseCase(repo).Execute(context.Background(),
		usecase.GrantWorkspaceRoleInput{WorkspaceID: kbWS, PrincipalID: kbPrincipal, Role: domain.GrantRoleAdmin})
	require.NoError(t, err)
	assert.Equal(t, domain.GrantRoleAdmin, wsGrant.Role)

	spGrant, err := usecase.NewGrantSpaceRoleUseCase(repo).Execute(context.Background(),
		usecase.GrantSpaceRoleInput{WorkspaceID: kbWS, SpaceID: kbSpace, PrincipalID: kbPrincipal, Role: domain.GrantRoleViewer})
	require.NoError(t, err)
	assert.Equal(t, domain.GrantRoleViewer, spGrant.Role)
}

func Test_権限剥奪_必須項目の検証(t *testing.T) {
	ctx := context.Background()
	repo := &mockKBPermissionRepo{}

	wsUC := usecase.NewRevokeWorkspaceRoleUseCase(repo)
	require.Error(t, wsUC.Execute(ctx, usecase.RevokeWorkspaceRoleInput{PrincipalID: kbPrincipal}))
	require.Error(t, wsUC.Execute(ctx, usecase.RevokeWorkspaceRoleInput{WorkspaceID: kbWS}))

	spUC := usecase.NewRevokeSpaceRoleUseCase(repo)
	require.Error(t, spUC.Execute(ctx, usecase.RevokeSpaceRoleInput{SpaceID: kbSpace, PrincipalID: kbPrincipal}))
	require.Error(t, spUC.Execute(ctx, usecase.RevokeSpaceRoleInput{WorkspaceID: kbWS, PrincipalID: kbPrincipal}))
	require.Error(t, spUC.Execute(ctx, usecase.RevokeSpaceRoleInput{WorkspaceID: kbWS, SpaceID: kbSpace}))
}

func Test_権限剥奪_repository_へ委譲する(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("DeleteWorkspaceGrant", mock.Anything, kbWS, kbPrincipal).Return(nil)
	repo.On("DeleteSpaceGrant", mock.Anything, kbWS, kbSpace, kbPrincipal).Return(nil)
	ctx := context.Background()

	require.NoError(t, usecase.NewRevokeWorkspaceRoleUseCase(repo).Execute(ctx,
		usecase.RevokeWorkspaceRoleInput{WorkspaceID: kbWS, PrincipalID: kbPrincipal}))
	require.NoError(t, usecase.NewRevokeSpaceRoleUseCase(repo).Execute(ctx,
		usecase.RevokeSpaceRoleInput{WorkspaceID: kbWS, SpaceID: kbSpace, PrincipalID: kbPrincipal}))
}

func Test_例外設定_ケイパビリティと向きの検証(t *testing.T) {
	uc := usecase.NewSetPageRestrictionUseCase(&mockKBPermissionRepo{})
	ctx := context.Background()

	_, err := uc.Execute(ctx, usecase.SetPageRestrictionInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
		Capability: domain.Capability("comment"), Mode: domain.RestrictionModeAllow,
	})
	require.ErrorIs(t, err, usecase.ErrInvalidCapability)

	_, err = uc.Execute(ctx, usecase.SetPageRestrictionInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
		Capability: domain.CapabilityView, Mode: domain.RestrictionMode("ignore"),
	})
	require.ErrorIs(t, err, usecase.ErrInvalidRestrictionMode)
}

func Test_例外設定_主体を確かめてから保存する(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbPrincipal).
		Return(&domain.Principal{ID: kbPrincipal, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	repo.On("UpsertPageRestriction", mock.Anything, kbWS, kbPage, kbPrincipal, domain.CapabilityView, domain.RestrictionModeDeny).
		Return(&domain.PageRestriction{
			WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
			Capability: domain.CapabilityView, Mode: domain.RestrictionModeDeny,
		}, nil)
	uc := usecase.NewSetPageRestrictionUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.SetPageRestrictionInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
		Capability: domain.CapabilityView, Mode: domain.RestrictionModeDeny,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.RestrictionModeDeny, got.Mode)
}

func Test_例外解除_検証と委譲(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("DeletePageRestriction", mock.Anything, kbWS, kbPage, kbPrincipal, domain.CapabilityView).Return(nil)
	uc := usecase.NewClearPageRestrictionUseCase(repo)
	ctx := context.Background()

	require.NoError(t, uc.Execute(ctx, usecase.ClearPageRestrictionInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Capability: domain.CapabilityView,
	}))
	require.ErrorIs(t, uc.Execute(ctx, usecase.ClearPageRestrictionInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Capability: domain.Capability("x"),
	}), usecase.ErrInvalidCapability)
	require.Error(t, uc.Execute(ctx, usecase.ClearPageRestrictionInput{
		PageID: kbPage, PrincipalID: kbPrincipal, Capability: domain.CapabilityView,
	}), "workspaceID 必須")
}
