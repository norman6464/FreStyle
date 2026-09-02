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
	repo.On("ListSpacePageViewFacts", mock.Anything, kbWS, kbSpace, uint64(1), false).
		Return([]repository.PageWithViewFacts{
			{Page: visible, Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)}},
			{Page: hidden, Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{HasAllowList: true},
			}},
		}, nil)
	uc := usecase.NewListViewablePagesUseCase(repo)

	out, err := uc.Execute(context.Background(), usecase.ListViewablePagesInput{
		WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1,
	})
	require.NoError(t, err)
	require.Len(t, out.Pages, 1)
	assert.Equal(t, "見える", out.Pages[0].Title)
	assert.True(t, out.HasHiddenChildren[usecase.HiddenChildrenRootKey],
		"落とした分は「在る」とだけ残す（枚数も題名も出さない）")
}

func Test_閲覧可能ページ一覧_見えない親の下は数えない(t *testing.T) {
	// 見える root ─ 見えない mid ─ その下に 2 枚（見えるものと見えないもの）。
	//
	// mid が見えないので、その配下は木に出ない（PageTreeOrphanHidden）。ここで mid の直下を
	// 数えてしまうと「見えない枝の中に何枚あるか」が漏れ、木から伏せた判断と食い違う。
	mid := "mid"
	rows := []repository.PageWithViewFacts{
		{
			Page:  domain.Page{ID: "root", WorkspaceID: kbWS, SpaceID: kbSpace, Title: "見える親"},
			Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)},
		},
		{
			Page: domain.Page{ID: mid, WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: strPtr("root"), Title: "隠れる中間"},
			Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{HasAllowList: true},
			},
		},
		{
			// 権限だけ見れば通るが、親が見えないので木には出ない（＝孤児）。
			Page:  domain.Page{ID: "orphan", WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: &mid, Title: "見えるが孤児"},
			Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)},
		},
		{
			Page: domain.Page{ID: "buried", WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: &mid, Title: "見えない孫"},
			Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{HasAllowList: true},
			},
		},
	}
	repo := &mockKBPermissionRepo{}
	repo.On("ListSpacePageViewFacts", mock.Anything, kbWS, kbSpace, uint64(1), false).Return(rows, nil)

	out, err := usecase.NewListViewablePagesUseCase(repo).
		Execute(context.Background(), usecase.ListViewablePagesInput{WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1})
	require.NoError(t, err)

	assert.True(t, out.HasHiddenChildren["root"], "見える親の直下で伏せた分は知らせる")
	assert.False(t, out.HasHiddenChildren[mid], "見えない親の直下は、伏せた孫が居ても知らせない")
	assert.False(t, out.HasHiddenChildren[usecase.HiddenChildrenRootKey])
	assert.Len(t, out.Pages, 2, "root と孤児。孤児を落とすのは木の組み立て側の役目")
}

func Test_閲覧可能ページ一覧_見える根が無いなら有無も返さない(t *testing.T) {
	// 根が非公開で、その子だけ閲覧できる形。
	//
	// 子は「見える」ので pages には入るが、親が見えないので木には繋がらず
	// （BuildPageTree の PageTreeOrphanHidden が落とす）、画面には 1 行も出ない。
	// このとき有無を返すと、応答が {"pages":[],"hasHiddenChildren":true} になり、
	// 存在しないスペースの {"pages":[],"hasHiddenChildren":false} と撃ち分けられる。
	root := domain.Page{ID: "root", WorkspaceID: kbWS, SpaceID: kbSpace, Title: "見えない根"}
	child := domain.Page{ID: "child", WorkspaceID: kbWS, SpaceID: kbSpace, ParentID: strPtr("root"), Title: "見える子"}

	repo := &mockKBPermissionRepo{}
	repo.On("ListSpacePageViewFacts", mock.Anything, kbWS, kbSpace, uint64(1), false).
		Return([]repository.PageWithViewFacts{
			{Page: root, Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{HasAllowList: true},
			}},
			{Page: child, Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)}},
		}, nil)

	out, err := usecase.NewListViewablePagesUseCase(repo).
		Execute(context.Background(), usecase.ListViewablePagesInput{WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1})
	require.NoError(t, err)

	assert.Len(t, out.Pages, 1, "子は見えるので一覧には入る（木から落とすのは組み立て側の役目）")
	assert.Empty(t, out.HasHiddenChildren, "画面に 1 行も出ないので、印は返さない")
}

func Test_閲覧可能ページ一覧_1件も見えないなら有無も返さない(t *testing.T) {
	// 存在しないスペースと「中身が 1 件も見えないスペース」を撃ち分けないための不変条件。
	// 有無を返すと、前者は false・後者は true になり、スペース ID の総当たりで実在が分かる。
	deny := domain.PageViewFacts{
		Role: kbGrantRole(domain.GrantRoleViewer),
		View: &domain.RestrictionFacts{HasAllowList: true},
	}
	repo := &mockKBPermissionRepo{}
	repo.On("ListSpacePageViewFacts", mock.Anything, kbWS, kbSpace, uint64(1), false).
		Return([]repository.PageWithViewFacts{
			{Page: domain.Page{ID: "p1", WorkspaceID: kbWS, SpaceID: kbSpace}, Facts: deny},
			{Page: domain.Page{ID: "p2", WorkspaceID: kbWS, SpaceID: kbSpace}, Facts: deny},
		}, nil)

	out, err := usecase.NewListViewablePagesUseCase(repo).
		Execute(context.Background(), usecase.ListViewablePagesInput{WorkspaceID: kbWS, SpaceID: kbSpace, UserID: 1})
	require.NoError(t, err)

	assert.Empty(t, out.Pages)
	assert.Empty(t, out.HasHiddenChildren, "存在しないスペースの応答と 1 バイトも変わらないこと")
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
	// 追加した瞬間から全員が書ける（既定 editor）。**無いときだけ**入れる口が呼ばれること
	// （上書きの Upsert だと、追加のやり直しで admin が editor に落ちる）。
	repo.On("GrantWorkspaceRoleIfAbsent", mock.Anything, kbWS, kbPrincipal, domain.GrantRoleEditor).
		Return(nil)
	uc := usecase.NewAddWorkspaceMemberUseCase(repo)

	got, err := uc.Execute(context.Background(), usecase.AddWorkspaceMemberInput{WorkspaceID: kbWS, UserID: 7})
	require.NoError(t, err)
	assert.Equal(t, domain.PrincipalKindUser, got.Kind)
	repo.AssertExpectations(t)

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

// strPtr は ParentID のようなポインタ項目をテストから書くための小道具。
func strPtr(v string) *string { return &v }

func Test_題名検索_見えないページを落とし件数を切る(t *testing.T) {
	visible := domain.Page{ID: "p-1", Title: "Docker 手順"}
	visible2 := domain.Page{ID: "p-2", Title: "Docker 入門"}
	denied := domain.Page{ID: "p-3", Title: "Docker 機密"}

	repo := &mockKBPermissionRepo{}
	repo.On("SearchWorkspacePageViewFacts", mock.Anything, "ws-1", uint64(7), "docker").
		Return([]repository.PageWithViewFacts{
			{Page: visible, Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)}},
			// 経路上で自分が deny されている行 — 一覧と同じ判定（ResolvePageView）で落ちること。
			{Page: denied, Facts: domain.PageViewFacts{
				Role: kbGrantRole(domain.GrantRoleViewer),
				View: &domain.RestrictionFacts{DeniedAnywhere: true},
			}},
			{Page: visible2, Facts: domain.PageViewFacts{Role: kbGrantRole(domain.GrantRoleViewer)}},
		}, nil)

	uc := usecase.NewSearchViewablePagesUseCase(repo)

	t.Run("deny された行は返らない", func(t *testing.T) {
		pages, err := uc.Execute(context.Background(), usecase.SearchViewablePagesInput{
			WorkspaceID: "ws-1", UserID: 7, Query: "docker",
		})
		require.NoError(t, err)
		require.Len(t, pages, 2)
		assert.Equal(t, "p-1", pages[0].ID)
		assert.Equal(t, "p-2", pages[1].ID)
	})

	t.Run("Limit は可視でふるった後に効き、範囲外は既定・上限へ畳まれる", func(t *testing.T) {
		// 入力 → 期待件数の表。可視は 2 件しか無いので、2 以上はすべて 2 になる。
		cases := []struct {
			name  string
			limit int
			want  int
		}{
			{name: "1 なら 1 件", limit: 1, want: 2 - 1},
			{name: "0 は既定 20 → 可視の全件", limit: 0, want: 2},
			{name: "負も既定 20 → 可視の全件", limit: -5, want: 2},
			{name: "上限 50 を超えても 50 に畳まれる（可視の全件）", limit: 999, want: 2},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				pages, err := uc.Execute(context.Background(), usecase.SearchViewablePagesInput{
					WorkspaceID: "ws-1", UserID: 7, Query: "docker", Limit: tc.limit,
				})
				require.NoError(t, err)
				require.Len(t, pages, tc.want)
				assert.Equal(t, "p-1", pages[0].ID, "並びは repo の返した順（題名順）を保つ")
			})
		}
	})
}

func Test_題名検索_空の問い合わせは誤り(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewSearchViewablePagesUseCase(repo)
	_, err := uc.Execute(context.Background(), usecase.SearchViewablePagesInput{
		WorkspaceID: "ws-1", UserID: 7, Query: "   ",
	})
	// 空で全件を返す口にしない（見えるページの全数が数えられる口になる）。
	assert.Error(t, err)
	repo.AssertNotCalled(t, "SearchWorkspacePageViewFacts", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ権限付与_必須項目と役割の検証(t *testing.T) {
	ctx := context.Background()
	uc := usecase.NewGrantPageRoleUseCase(&mockKBPermissionRepo{})

	_, err := uc.Execute(ctx, usecase.GrantPageRoleInput{
		PageID: kbPage, PrincipalID: kbPrincipal, Role: domain.GrantRoleEditor,
	})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.GrantPageRoleInput{
		WorkspaceID: kbWS, PrincipalID: kbPrincipal, Role: domain.GrantRoleEditor,
	})
	require.Error(t, err, "pageID 必須")
	_, err = uc.Execute(ctx, usecase.GrantPageRoleInput{
		WorkspaceID: kbWS, PageID: kbPage, Role: domain.GrantRoleEditor,
	})
	require.Error(t, err, "principalID 必須")
	_, err = uc.Execute(ctx, usecase.GrantPageRoleInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Role: domain.GrantRole("owner"),
	})
	require.ErrorIs(t, err, usecase.ErrInvalidGrantRole)
}

func Test_ページ権限付与_別ワークスペースの主体は拒否(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbPrincipal).Return(nil, repository.ErrPrincipalNotFound)

	_, err := usecase.NewGrantPageRoleUseCase(repo).Execute(context.Background(), usecase.GrantPageRoleInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Role: domain.GrantRoleEditor,
	})
	require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
	// 主体を確かめる前に書き込まないこと（FK 違反ではなく not found として返すため）。
	repo.AssertNotCalled(t, "UpsertPageGrant",
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ページ権限付与_repository_へ委譲する(t *testing.T) {
	repo := &mockKBPermissionRepo{}
	repo.On("FindPrincipal", mock.Anything, kbWS, kbPrincipal).
		Return(&domain.Principal{ID: kbPrincipal, WorkspaceID: kbWS, Kind: domain.PrincipalKindUser}, nil)
	repo.On("UpsertPageGrant", mock.Anything, kbWS, kbPage, kbPrincipal, domain.GrantRoleAdmin).
		Return(&domain.PageGrant{
			WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Role: domain.GrantRoleAdmin,
		}, nil)

	got, err := usecase.NewGrantPageRoleUseCase(repo).Execute(context.Background(), usecase.GrantPageRoleInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal, Role: domain.GrantRoleAdmin,
	})
	require.NoError(t, err)
	assert.Equal(t, domain.GrantRoleAdmin, got.Role)
	assert.Equal(t, kbPage, got.PageID)
}

func Test_ページ権限剥奪_必須項目の検証と委譲(t *testing.T) {
	ctx := context.Background()
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewRevokePageRoleUseCase(repo)

	require.Error(t, uc.Execute(ctx, usecase.RevokePageRoleInput{PageID: kbPage, PrincipalID: kbPrincipal}))
	require.Error(t, uc.Execute(ctx, usecase.RevokePageRoleInput{WorkspaceID: kbWS, PrincipalID: kbPrincipal}))
	require.Error(t, uc.Execute(ctx, usecase.RevokePageRoleInput{WorkspaceID: kbWS, PageID: kbPage}))

	repo.On("DeletePageGrant", mock.Anything, kbWS, kbPage, kbPrincipal).Return(nil)
	require.NoError(t, uc.Execute(ctx, usecase.RevokePageRoleInput{
		WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal,
	}))
}

func Test_ページ権限一覧_必須項目の検証と委譲(t *testing.T) {
	ctx := context.Background()
	repo := &mockKBPermissionRepo{}
	uc := usecase.NewListPageGrantsUseCase(repo)

	_, err := uc.Execute(ctx, usecase.ListPageGrantsInput{PageID: kbPage})
	require.Error(t, err, "workspaceID 必須")
	_, err = uc.Execute(ctx, usecase.ListPageGrantsInput{WorkspaceID: kbWS})
	require.Error(t, err, "pageID 必須")

	repo.On("ListPageGrants", mock.Anything, kbWS, kbPage).
		Return([]domain.PageGrant{{WorkspaceID: kbWS, PageID: kbPage, PrincipalID: kbPrincipal}}, nil)
	got, err := uc.Execute(ctx, usecase.ListPageGrantsInput{WorkspaceID: kbWS, PageID: kbPage})
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, kbPrincipal, got[0].PrincipalID)
}
