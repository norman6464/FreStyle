//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scopeOf は入れ物（スペース）1 つの実効権限を解いて返す。
// 事実の収集（SQL）→ 規則の適用（domain）の 2 段を、本番と同じ順で通す小道具。
func scopeOf(ctx context.Context, t *testing.T, f kbPermFixture, spaceID string, userID uint64) domain.ScopePermission {
	t.Helper()
	facts, err := f.perm.SpacePermissionFactsForUser(ctx, f.ws, spaceID, userID)
	require.NoError(t, err)
	return domain.ResolveScopePermission(*facts)
}

// workspaceScopeOf はワークスペースそのものの実効権限を解いて返す。
func workspaceScopeOf(ctx context.Context, t *testing.T, f kbPermFixture, userID uint64) domain.ScopePermission {
	t.Helper()
	facts, err := f.perm.WorkspacePermissionFactsForUser(ctx, f.ws, userID)
	require.NoError(t, err)
	return domain.ResolveScopePermission(*facts)
}

// TestKnowledgeBaseScopePermission_Integration はページを介さない権限照会
// （スペース / ワークスペース単位）を実 PostgreSQL で固定する。
//
// この口は「対象がまだ存在しない操作」（空のスペースへの最初のページ作成 / スペースの作成）
// のためにあり、ページ単位の例外を見ない。見ないこと自体が正しい設計だが、
// 見ないまま**ページを名指しする操作**に使うと必ず緩い側へ倒れるので、
// 「ページ単位の答えと食い違わないこと」と「食い違ってよい範囲」の両方をここで固定する。
func TestKnowledgeBaseScopePermission_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("非メンバーは何もできない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)

		assert.Equal(t, domain.ScopePermission{}, scopeOf(ctx, t, f, f.spaceA, f.alice),
			"principal が無い相手には役割が 1 つも届かない")
		assert.Equal(t, domain.ScopePermission{}, workspaceScopeOf(ctx, t, f, f.alice))
	})

	t.Run("所属しただけで役割が無ければ何もできない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)

		assert.Equal(t, domain.ScopePermission{}, scopeOf(ctx, t, f, f.spaceA, f.alice),
			"所属は入口であって権限ではない（grant が無ければ空）")
	})

	t.Run("ワークスペースのgrantは配下の全スペースに届く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		alice := f.principalFor(ctx, t, f.alice)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		for _, space := range []string{f.spaceA, f.spaceB} {
			perm := scopeOf(ctx, t, f, space, f.alice)
			assert.True(t, perm.CanEdit, "grant を張っていないスペースにも届く")
			assert.False(t, perm.CanManage)
		}
	})

	t.Run("スペースとワークスペースの強い方が効く", func(t *testing.T) {
		cases := []struct {
			name      string
			workspace domain.GrantRole
			space     domain.GrantRole
			wantEdit  bool
			wantAdmin bool
		}{
			{name: "スペースの方が強い", workspace: domain.GrantRoleViewer, space: domain.GrantRoleEditor, wantEdit: true},
			{name: "ワークスペースの方が強い", workspace: domain.GrantRoleAdmin, space: domain.GrantRoleViewer, wantEdit: true, wantAdmin: true},
			{name: "どちらも弱い", workspace: domain.GrantRoleViewer, space: domain.GrantRoleCommenter},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := setupKBPermission(t, sqlDB)
				alice := f.principalFor(ctx, t, f.alice)
				_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, tc.workspace)
				require.NoError(t, err)
				f.grantSpace(ctx, t, f.spaceA, alice.ID, tc.space)

				perm := scopeOf(ctx, t, f, f.spaceA, f.alice)
				assert.True(t, perm.CanView)
				assert.Equal(t, tc.wantEdit, perm.CanEdit,
					"弱い方を採るとスペースに viewer を張るだけでワークスペース管理者を締め出せてしまう")
				assert.Equal(t, tc.wantAdmin, perm.CanManage)
			})
		}
	})

	t.Run("グループ経由の役割も届く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "開発チーム")
		require.NoError(t, err)
		require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, alice.ID))
		f.grantSpace(ctx, t, f.spaceA, group.ID, domain.GrantRoleEditor)

		assert.True(t, scopeOf(ctx, t, f, f.spaceA, f.alice).CanEdit, "グループの役割が本人に届く")
		assert.False(t, scopeOf(ctx, t, f, f.spaceA, f.bob).CanEdit, "グループ外には届かない")
		assert.False(t, scopeOf(ctx, t, f, f.spaceB, f.alice).CanEdit, "別スペースには届かない")

		// 抜けた瞬間に効かなくなる。
		require.NoError(t, f.perm.RemoveGroupMember(ctx, f.ws, group.ID, alice.ID))
		assert.False(t, scopeOf(ctx, t, f, f.spaceA, f.alice).CanEdit)
	})

	t.Run("グループ経由の役割はワークスペース単位の判定にも届く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		alice := f.principalFor(ctx, t, f.alice)
		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "管理チーム")
		require.NoError(t, err)
		require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, alice.ID))
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, group.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		assert.True(t, workspaceScopeOf(ctx, t, f, f.alice).CanManage)
		assert.False(t, workspaceScopeOf(ctx, t, f, f.bob).CanManage)
	})

	t.Run("スペース全員の主体経由の役割も届く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)
		everyone := f.everyoneOf(ctx, t, f.spaceA)
		f.grantSpace(ctx, t, f.spaceA, everyone.ID, domain.GrantRoleEditor)

		assert.True(t, scopeOf(ctx, t, f, f.spaceA, f.alice).CanEdit, "そのスペースの全員に届く")
		assert.False(t, scopeOf(ctx, t, f, f.spaceB, f.alice).CanEdit, "別スペースの全員には届かない")

		// 非メンバーには届かない。「全員」はワークスペースのメンバーの中の全員という意味。
		assert.False(t, scopeOf(ctx, t, f, f.spaceA, f.bob).CanEdit,
			"所属していない相手にスペース全員の役割が届いてはいけない")
	})

	t.Run("スペース全員の主体はワークスペース単位の判定には数えない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)
		everyone := f.everyoneOf(ctx, t, f.spaceA)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, everyone.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		assert.False(t, workspaceScopeOf(ctx, t, f, f.alice).CanManage,
			"どこか 1 つのスペースの全員に張った grant がテナント全体の権限に化けてはいけない")
		assert.True(t, scopeOf(ctx, t, f, f.spaceA, f.alice).CanManage,
			"そのスペースの中では効く（届く先はスペースに閉じる）")
	})

	t.Run("別テナントのスペースIDでは役割を返さない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		alice := f.principalFor(ctx, t, f.alice)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		// f.otherSpc は別ワークスペースのスペース。実在は確かめるが、このテナントのものではない。
		_, err = f.perm.SpacePermissionFactsForUser(ctx, f.ws, f.otherSpc, f.alice)
		assert.ErrorIs(t, err, repository.ErrSpaceNotFound,
			"実在を確かめずに役割だけ集めると、ワークスペースの grant が他テナントのスペースに届いてしまう")

		_, err = f.perm.SpacePermissionFactsForUser(ctx, f.ws, newID(), f.alice)
		assert.ErrorIs(t, err, repository.ErrSpaceNotFound, "存在しないスペースも同じ扱い")

		_, err = f.perm.SpacePermissionFactsForUser(ctx, f.ws, "not-a-uuid", f.alice)
		assert.ErrorIs(t, err, repository.ErrSpaceNotFound, "形が UUID でない ID も同じ扱い")
	})

	t.Run("スペース単位とページ単位の答えは例外が無ければ一致する", func(t *testing.T) {
		// スペースの判定（役割の集合を domain が畳む）とページの判定（SQL が強さを返す）は
		// 実装が別なので、同じ既定に対して同じ答えになることを役割ごとに固定する。
		// ここが割れると「ページは編集できるのに直下に作れない」（逆も）になる。
		for _, role := range domain.ValidGrantRoles {
			t.Run(string(role), func(t *testing.T) {
				f := setupKBPermission(t, sqlDB)
				alice := f.principalFor(ctx, t, f.alice)
				f.grantSpace(ctx, t, f.spaceA, alice.ID, role)
				page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")

				scope := scopeOf(ctx, t, f, f.spaceA, f.alice)
				pagePerm := f.permFor(ctx, t, page.ID, f.alice)

				assert.Equal(t, pagePerm.CanView, scope.CanView, "閲覧の答えが経路で割れている")
				assert.Equal(t, pagePerm.CanEdit, scope.CanEdit, "編集の答えが経路で割れている")
			})
		}
	})

	t.Run("スペース単位の答えはページの例外を見ない", func(t *testing.T) {
		// この口の限界をそのまま固定する。ページに deny があってもスペースの既定は editor のまま
		// 返る（例外の層を集めていないため）。だから**ページを名指しする操作に使ってはいけない**。
		// 呼び出し側がそれを守っていることは handler の結合テストが確かめる。
		f := setupKBPermission(t, sqlDB)
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleEditor)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		f.restrict(ctx, t, page.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeDeny)

		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanView, "ページ単位では deny が効く")
		assert.True(t, scopeOf(ctx, t, f, f.spaceA, f.alice).CanEdit,
			"スペース単位はページの例外を見ない（見ていない事実を答えに混ぜないための設計）")
	})
}

// TestKnowledgeBaseMemberWorkspaces_Integration は所属ワークスペース一覧を実 PostgreSQL で固定する。
// ナレッジ基盤で唯一テナントを跨いで読む口なので、絞り込みが緩むと全テナントが漏れる。
func TestKnowledgeBaseMemberWorkspaces_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("所属しているものだけをslug順で返す", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// slug が昇順にならない順で所属させ、並びがクエリ側で決まることを見る。
		_, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		got, err := f.perm.ListMemberWorkspaces(ctx, f.alice)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, "perm-main", got[0].Slug)
		assert.Equal(t, "perm-other", got[1].Slug)
	})

	t.Run("所属していないワークスペースは漏らさない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		_, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		// bob は別テナントだけ。alice の一覧に混ざってはいけない。
		_, err = f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.bob)
		require.NoError(t, err)

		got, err := f.perm.ListMemberWorkspaces(ctx, f.alice)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "perm-main", got[0].Slug)

		none, err := f.perm.ListMemberWorkspaces(ctx, f.carol)
		require.NoError(t, err)
		assert.Empty(t, none, "どこにも所属していなければ 0 件")
	})

	t.Run("ユーザー以外の主体は所属として数えない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// グループやスペース全員の主体は「誰かの所属」ではない。
		_, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "開発チーム")
		require.NoError(t, err)
		_, err = f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)

		got, err := f.perm.ListMemberWorkspaces(ctx, f.alice)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("所属を消すと一覧から消える", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		principal, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		require.NoError(t, f.perm.DeletePrincipal(ctx, f.ws, principal.ID))

		got, err := f.perm.ListMemberWorkspaces(ctx, f.alice)
		require.NoError(t, err)
		assert.Empty(t, got, "所属は principals の行が唯一の表現")
	})
}
