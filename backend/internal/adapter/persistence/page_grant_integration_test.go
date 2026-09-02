//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ページ単位の付与（page_grants）が、既定の役割としてどう効くかを実 PostgreSQL で固定する。
//
// 付与は「足す」層しかない。3 段（ワークスペース / スペース / ページ）から届いた役割のうち
// 最も強いものが実効になり、下の段が上の段を弱めることはない。ここで確かめたいのは、
// その足し算と、経路（自分と祖先）をどう辿るか。
//
// **1 ページの解決と一覧の解決が割れないこと**を特に見る。既定の役割を計算する箇所が
// 6 つに分かれており、1 つでも枝を足し忘れると「開くと編集できるのに一覧では読み取り
// 専用に見える」というずれになる。経路ごとに別の答えを返すのが、この実装で最も起こり
// やすい壊れ方。

// grantPage はページに既定の役割を 1 行張る（grantSpace のページ版）。
//
// SQL を直に書かず repository を通すのは、読みの検証が書き経路と同じ道を通るようにするため。
// 直に入れると、書き経路が壊れていても読みのテストだけは緑のままになる。
func (f kbPermFixture) grantPage(ctx context.Context, t *testing.T, pageID, principalID string, role domain.GrantRole) {
	t.Helper()
	_, err := f.perm.UpsertPageGrant(ctx, f.ws, pageID, principalID, role)
	require.NoError(t, err)
}

func TestPageGrant_付与した本人だけに既定の役割が足される_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
	alice := f.principalFor(ctx, t, f.alice)
	f.principalFor(ctx, t, f.bob)

	// 付与の前は、所属しているだけでは何もできない（grant が無ければ強さ 0）。
	require.False(t, f.permFor(ctx, t, page, f.alice).CanView, "付与前は見えない")

	f.grantPage(ctx, t, page, alice.ID, domain.GrantRoleEditor)

	got := f.permFor(ctx, t, page, f.alice)
	assert.True(t, got.CanView, "付与した本人は見える")
	assert.True(t, got.CanEdit, "editor なので編集もできる")

	// 足し算なので、張っていない相手の権限は動かない。
	other := f.permFor(ctx, t, page, f.bob)
	assert.False(t, other.CanView, "他人の権限は変わらない")
	assert.False(t, other.CanEdit)
}

func TestPageGrant_祖先に張ると子孫へ降りる_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	grandchild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &child, "孫").ID
	alice := f.principalFor(ctx, t, f.alice)

	// 親にだけ張る。
	f.grantPage(ctx, t, parent, alice.ID, domain.GrantRoleEditor)

	for _, c := range []struct {
		name string
		page string
	}{
		{"親", parent},
		{"子", child},
		{"孫", grandchild},
	} {
		got := f.permFor(ctx, t, c.page, f.alice)
		assert.True(t, got.CanEdit, "%s は編集できる（親に張った付与が降りる）", c.name)
	}
}

func TestPageGrant_子に張っても親へは上がらない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	alice := f.principalFor(ctx, t, f.alice)

	f.grantPage(ctx, t, child, alice.ID, domain.GrantRoleEditor)

	assert.True(t, f.permFor(ctx, t, child, f.alice).CanEdit, "張った子は編集できる")
	assert.False(t, f.permFor(ctx, t, parent, f.alice).CanView,
		"親へは上がらない（経路は祖先だけを辿る）")
}

func TestPageGrant_強い方が採られる_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
	alice := f.principalFor(ctx, t, f.alice)

	// スペースに viewer、ページに editor。強い方（editor）が採られる。
	f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
	f.grantPage(ctx, t, page, alice.ID, domain.GrantRoleEditor)

	got := f.permFor(ctx, t, page, f.alice)
	assert.True(t, got.CanEdit, "ページの editor が採られる")

	// 逆向き（スペースが強い）でも、弱い付与が降格させないことを見る。
	f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleAdmin)
	f.grantPage(ctx, t, page, alice.ID, domain.GrantRoleViewer)
	assert.True(t, f.permFor(ctx, t, page, f.alice).CanEdit,
		"ページに viewer を張ってもスペースの admin は下がらない（付与は足し算だけ）")
}

// 付与を剥がすと届かなくなることを、読みの経路まで通して見る。
//
// **見えなくする手段はこれだけ**（同じスペースの中で 1 枚だけ隠す層は無い）。
// 剥がしたのに読みの経路がまだ役割を返す、という壊れ方をここで捕まえる。
func TestPageGrant_剥がすと届かなくなる_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	alice := f.principalFor(ctx, t, f.alice)

	f.grantPage(ctx, t, parent, alice.ID, domain.GrantRoleAdmin)
	require.True(t, f.permFor(ctx, t, child, f.alice).CanEdit, "前提: 祖先の付与で編集できている")

	require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, parent, alice.ID))

	got := f.permFor(ctx, t, child, f.alice)
	assert.False(t, got.CanView, "剥がせば届かない（ほかに届く段が無いので何もできない）")
	assert.False(t, got.CanEdit)
	assert.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.alice), "一覧からも消える")
}

func TestPageGrant_一覧と1ページの解決が一致する_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	outside := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象外").ID
	alice := f.principalFor(ctx, t, f.alice)

	f.grantPage(ctx, t, parent, alice.ID, domain.GrantRoleEditor)

	// 一覧（1 クエリでスペース全体の事実を集める経路）。
	listed := f.viewablePageIDs(ctx, t, f.spaceA, f.alice)

	// 1 ページずつの解決と突き合わせる。ここが割れると、開けるのに一覧に出ない
	// （またはその逆）というずれになる。
	for _, c := range []struct {
		name    string
		page    string
		canView bool
	}{
		{"親（付与を張った）", parent, true},
		{"子（親から降りる）", child, true},
		{"対象外（付与の経路に無い）", outside, false},
	} {
		single := f.permFor(ctx, t, c.page, f.alice).CanView
		assert.Equal(t, c.canView, single, "%s: 1 ページの解決", c.name)
		assert.Equal(t, c.canView, contains(listed, c.page), "%s: 一覧の結果が 1 ページの解決と一致する", c.name)
	}
}

// スペース全員（space_all）宛ての付与も、どの経路でも同じ答えになることを固定する。
//
// 検索と ID 指定の解決だけは「自分と所属グループ」と「スペース全員」を別々の CTE で
// 持っており、片方だけを見て付与を評価すると、開けるのに検索に出ないページができる。
func TestPageGrant_スペース全員宛ての付与が全経路で一致する_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "全員に配る").ID
	// alice はワークスペースに所属しているだけ（個人の付与は持たない）。
	f.principalFor(ctx, t, f.alice)
	everyone := f.everyoneOf(ctx, t, f.spaceA)

	f.grantPage(ctx, t, page, everyone.ID, domain.GrantRoleEditor)

	// 1 ページの解決では見える。
	require.True(t, f.permFor(ctx, t, page, f.alice).CanView, "前提: 直接開けば見える")

	// 検索でも出る。
	found, err := usecase.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
		usecase.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: f.alice, Query: "全員"})
	require.NoError(t, err)
	assert.True(t, containsPageID(found, page), "検索でも出る（開けるのに検索に出ないのは経路のずれ）")

	// ID 指定でも同じ。
	rows, err := f.perm.ListWorkspacePageViewFactsByIDs(ctx, f.ws, f.alice, []string{page})
	require.NoError(t, err)
	require.Len(t, rows, 1, "ID 指定でも 1 行返る")
	assert.True(t, domain.ResolvePageView(rows[0].Role), "ID 指定でも見える")
}

func containsPageID(pages []domain.Page, id string) bool {
	for _, p := range pages {
		if p.ID == id {
			return true
		}
	}
	return false
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

func TestPageGrant_書き経路_付与と取り消し_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
	alice := f.principalFor(ctx, t, f.alice)

	t.Run("張った内容がそのまま返る", func(t *testing.T) {
		got, err := f.perm.UpsertPageGrant(ctx, f.ws, page, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		assert.Equal(t, page, got.PageID)
		assert.Equal(t, alice.ID, got.PrincipalID)
		assert.Equal(t, domain.GrantRoleEditor, got.Role)
	})

	t.Run("同じ主体に 2 度張ると行は増えず役割だけ変わる", func(t *testing.T) {
		before, err := f.perm.UpsertPageGrant(ctx, f.ws, page, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		after, err := f.perm.UpsertPageGrant(ctx, f.ws, page, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.Equal(t, domain.GrantRoleViewer, after.Role)
		assert.Equal(t, before.CreatedAt, after.CreatedAt, "作成時刻は動かない")
		assert.True(t, after.UpdatedAt.After(before.UpdatedAt) || after.UpdatedAt.Equal(before.UpdatedAt),
			"更新時刻は巻き戻らない")

		rows, err := f.perm.ListPageGrants(ctx, f.ws, page)
		require.NoError(t, err)
		assert.Len(t, rows, 1, "行は 1 つのまま")
	})

	t.Run("取り消しは冪等", func(t *testing.T) {
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, page, alice.ID))
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, page, alice.ID), "元から無くても成功する")

		rows, err := f.perm.ListPageGrants(ctx, f.ws, page)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("UUID として読めない ID は not found（取り消しは黙って成功）", func(t *testing.T) {
		_, err := f.perm.UpsertPageGrant(ctx, f.ws, page, "not-a-uuid", domain.GrantRoleEditor)
		assert.ErrorIs(t, err, repository.ErrPrincipalNotFound)

		assert.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, page, "not-a-uuid"))

		rows, err := f.perm.ListPageGrants(ctx, f.ws, "not-a-uuid")
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}

func TestPageGrant_書き経路_テナントと参照の整合_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
	alice := f.principalFor(ctx, t, f.alice)

	// 張れないページの一覧。どれも複合 FK（workspace_id, page_id）が塞ぐ。
	// ここが 1 つでも通ると、自分のワークスペースの主体に他テナントのページの権限を配れる。
	for _, c := range []struct {
		name   string
		pageID func() string
	}{
		{"別ワークスペースのページ", func() string {
			return mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "他社").ID
		}},
		{"存在しないページ", newID},
	} {
		t.Run(c.name+"には張れない", func(t *testing.T) {
			target := c.pageID()
			_, err := f.perm.UpsertPageGrant(ctx, f.ws, target, alice.ID, domain.GrantRoleEditor)
			require.Error(t, err)

			// 断られただけでなく、行が 1 つも残っていないこと。
			// 「エラーは返るが書けている」は、次の一覧で初めて見つかる壊れ方になる。
			var count int
			require.NoError(t, sqlDB.QueryRow(
				`SELECT count(*) FROM page_grants WHERE page_id = $1`, target,
			).Scan(&count))
			assert.Zero(t, count, "断った要求の行を残さない")
		})
	}

	t.Run("ページを消すと張った行も消える", func(t *testing.T) {
		doomed := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "消す").ID
		f.grantPage(ctx, t, doomed, alice.ID, domain.GrantRoleEditor)

		_, err := sqlDB.Exec(`DELETE FROM pages WHERE workspace_id = $1 AND id = $2`, f.ws, doomed)
		require.NoError(t, err)

		rows, err := f.perm.ListPageGrants(ctx, f.ws, doomed)
		require.NoError(t, err)
		assert.Empty(t, rows, "ページと一緒に消える（孤児の grant を残さない）")
	})

	t.Run("主体を消すと張った行も消える", func(t *testing.T) {
		f.grantPage(ctx, t, page, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, f.perm.DeletePrincipal(ctx, f.ws, alice.ID))

		rows, err := f.perm.ListPageGrants(ctx, f.ws, page)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})
}

func TestPageGrant_一覧はその段に張った行だけを返す_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	alice := f.principalFor(ctx, t, f.alice)

	f.grantPage(ctx, t, parent, alice.ID, domain.GrantRoleEditor)

	// 解決は祖先まで辿るので、子でも編集できる。
	require.True(t, f.permFor(ctx, t, child, f.alice).CanEdit, "前提: 親の付与が子へ降りている")

	// 一覧はそれを写さない。「どの段で足したか」が分からなくなると、
	// 取り消すべき行を人が選べない（子で消そうとしても消せる行がそこには無い）。
	onChild, err := f.perm.ListPageGrants(ctx, f.ws, child)
	require.NoError(t, err)
	assert.Empty(t, onChild, "祖先の行は含めない")

	onParent, err := f.perm.ListPageGrants(ctx, f.ws, parent)
	require.NoError(t, err)
	require.Len(t, onParent, 1)
	assert.Equal(t, alice.ID, onParent[0].PrincipalID)
}

// ページ単位の付与が無いとき、権限操作の入口の答えがスペースを名指しで引いたときと
// 一致することを固定する。
//
// # なぜこれが要るのか
//
// 権限操作の入口（kbPermissionGate.requirePageAdmin）は、以前はスペースを名指しで引いた
// 役割だけを見ていた。ページ単位の付与を数に入れるためにページ経由の解決へ差し替えたが、
// **その差し替えで意図しない方向に緩む余地がある**。2 つの経路は別々の SQL で、
// 主体の集め方（自分 / 所属グループ / スペース全員）も private スペースの扱いも
// それぞれに書いてある。片方だけが緩ければ、ページ単位の付与を 1 行も張っていないのに
// 「スペースでは admin ではないのに、ページの権限は変えられる」人ができる。
//
// 増やしたかったのは「ページに admin を張られた人」だけ。それ以外は前と同じであること。
func TestPageGrant_付与が無ければ入口の答えはスペース経由と一致する_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	for _, c := range []struct {
		name  string
		setup func(t *testing.T, f kbPermFixture, spaceID string)
		// private はスペースを private にしてから配役する（ワークスペースの既定が届かない段）。
		private bool
		// wantManage はその配役で権限を変えられるべきか。
		//
		// 2 つの経路が一致することだけを見ると、**両方とも同じように壊れた場合に空振りする**
		// （どちらも false を返すようになれば、一致はするが誰も権限を変えられない）。
		// 期待値を先に書いて、一致と正しさの両方を見る。
		wantManage bool
	}{
		{name: "何も張らない", setup: func(*testing.T, kbPermFixture, string) {}, wantManage: false},
		{
			name:       "ワークスペースに admin",
			wantManage: true,
			setup: func(t *testing.T, f kbPermFixture, _ string) {
				_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleAdmin)
				require.NoError(t, err)
			},
		},
		{
			name:       "スペースに admin",
			wantManage: true,
			setup: func(t *testing.T, f kbPermFixture, spaceID string) {
				f.grantSpace(ctx, t, spaceID, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleAdmin)
			},
		},
		{
			name:       "スペースに editor（admin には届かない）",
			wantManage: false,
			setup: func(t *testing.T, f kbPermFixture, spaceID string) {
				f.grantSpace(ctx, t, spaceID, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleEditor)
			},
		},
		{
			name:       "スペース全員に admin",
			wantManage: true,
			setup: func(t *testing.T, f kbPermFixture, spaceID string) {
				f.principalFor(ctx, t, f.alice)
				f.grantSpace(ctx, t, spaceID, f.everyoneOf(ctx, t, spaceID).ID, domain.GrantRoleAdmin)
			},
		},
		{
			name:       "所属グループに admin",
			wantManage: true,
			setup: func(t *testing.T, f kbPermFixture, spaceID string) {
				self := f.principalFor(ctx, t, f.alice)
				group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "運用")
				require.NoError(t, err)
				require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, self.ID))
				f.grantSpace(ctx, t, spaceID, group.ID, domain.GrantRoleAdmin)
			},
		},
		{
			name:       "private スペースにワークスペースの admin（届かない）",
			wantManage: false,
			private:    true,
			setup: func(t *testing.T, f kbPermFixture, _ string) {
				_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleAdmin)
				require.NoError(t, err)
			},
		},
		{
			name:       "private スペースにスペースの admin（届く）",
			wantManage: true,
			private:    true,
			setup: func(t *testing.T, f kbPermFixture, spaceID string) {
				f.grantSpace(ctx, t, spaceID, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleAdmin)
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := setupKBPermission(t, sqlDB)
			if c.private {
				f.makePrivate(t, f.spaceA)
			}
			page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
			c.setup(t, f, f.spaceA)

			// 入口が実際に見る答え（ページ経由）。
			facts, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, page, f.alice)
			require.NoError(t, err)
			viaPage := domain.ResolvePagePermission(*facts).CanManage

			// 差し替える前に見ていた答え（スペースを名指し）。
			scope, err := f.perm.SpacePermissionFactsForUser(ctx, f.ws, f.spaceA, f.alice)
			require.NoError(t, err)
			viaSpace := domain.ResolveScopePermission(*scope).CanManage

			assert.Equal(t, c.wantManage, viaPage, "ページ経由の答え")
			assert.Equal(t, viaSpace, viaPage,
				"ページ単位の付与が 1 行も無いのに、2 つの経路で答えが割れている")
		})
	}
}

// 権限を張れる相手の一覧が、表示名を正しい出どころから引くことを確かめる。
//
// 名前の正本は kind ごとに別の表にある（group は principals、user は users、
// space_all は spaces）。1 箇所でも取り違えると、画面に UUID や空欄が並んで
// 相手を選べなくなる。
func TestGrantablePrincipals_表示名を正しい出どころから引く_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	alice := f.principalFor(ctx, t, f.alice)
	group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "開発チーム")
	require.NoError(t, err)
	everyone := f.everyoneOf(ctx, t, f.spaceA)

	// users.name は createUser が渡した接頭辞、spaces.name は createSpace の key と同じ。
	var aliceName, spaceName string
	require.NoError(t, sqlDB.QueryRow(`SELECT name FROM users WHERE id = $1`, f.alice).Scan(&aliceName))
	require.NoError(t, sqlDB.QueryRow(
		`SELECT name FROM spaces WHERE workspace_id = $1 AND id = $2`, f.ws, f.spaceA,
	).Scan(&spaceName))

	got, err := f.perm.ListGrantablePrincipals(ctx, f.ws)
	require.NoError(t, err)

	byID := map[string]domain.GrantablePrincipal{}
	for _, p := range got {
		byID[p.ID] = p
	}

	assert.Equal(t, aliceName, byID[alice.ID].Name, "ユーザーの名前は users から引く")
	assert.Equal(t, domain.PrincipalKindUser, byID[alice.ID].Kind)
	assert.Equal(t, "開発チーム", byID[group.ID].Name, "グループの名前は principals から引く")
	assert.Equal(t, spaceName, byID[everyone.ID].Name, "スペース全員の名前は spaces から引く")
}

func TestGrantablePrincipals_選べない相手と他テナントを混ぜない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	f.principalFor(ctx, t, f.alice)
	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "共有元").ID

	// 共有リンクを 1 本発行する。来訪者を表す主体（kind='share_link'）が一緒に作られる。
	link, err := f.perm.CreateShareLink(ctx, repository.ShareLinkWrite{
		WorkspaceID:     f.ws,
		PageID:          page,
		Capability:      domain.CapabilityView,
		TokenHash:       []byte("0123456789abcdef0123456789abcdef"),
		CreatedByUserID: f.alice,
	})
	require.NoError(t, err)
	require.NotEmpty(t, link.ID)

	// 別テナントにも主体を作る。
	otherPrincipal, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.bob)
	require.NoError(t, err)

	got, err := f.perm.ListGrantablePrincipals(ctx, f.ws)
	require.NoError(t, err)

	for _, p := range got {
		assert.NotEqual(t, domain.PrincipalKindShareLink, p.Kind,
			"リンクの来訪者は選ぶ相手ではない（役割を与えても意味を持たない）")
		assert.NotEqual(t, otherPrincipal.ID, p.ID, "他テナントの主体を混ぜない")
	}
	// 空振り防止: 選べる相手自体はちゃんと返っている。
	assert.NotEmpty(t, got)
}

func TestGrantablePrincipals_名前が引けなくても行を落とさない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	// 名前を空にしたユーザーの主体（users.name が空文字の行）。
	// 一覧から黙って消えると、この相手に張った権限が画面に出たまま選べなくなる。
	blank := f.principalFor(ctx, t, f.carol)
	_, err := sqlDB.Exec(`UPDATE users SET name = '' WHERE id = $1`, f.carol)
	require.NoError(t, err)

	got, err := f.perm.ListGrantablePrincipals(ctx, f.ws)
	require.NoError(t, err)

	found := false
	for _, p := range got {
		if p.ID == blank.ID {
			found = true
			assert.Empty(t, p.Name, "名前は空のまま返す（作り話をしない）")
		}
	}
	assert.True(t, found, "名前が無くても行は残る")
}

// 木を下るほど役割が弱くならないことを、**実 PostgreSQL のクエリで**確かめる。
//
// これが要るのは、domain 側の単調性（StrongestGrantRole のテスト）だけでは証明にならないため。
// ページ 1 枚 / 一覧の役割は SQL の `GREATEST(...)` が畳んだ強さを persistence が戻して作るので、
// domain の合成関数を一度も通らない。両方が単調でなければ「親は編集できるが子は編集できない」が
// 起きないとは言えない。
//
// この性質は 2 つの前提に乗っている。どちらも DB が強制する:
//   - 子孫の経路は祖先の経路を含む（page_paths が closure）
//   - スペースは親子で揃う（fk_pages_parent が (workspace_id, space_id, parent_id) を参照する）
//
// 壊れ方は「経路の辿り方を取り違える」形で起きる。祖先ではなく子孫を集めてしまうと、
// 深いページほど弱くなる。根だけ見て通す実装では気づけない。
func TestPageGrant_木を下るほど役割は弱くならない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根").ID
	mid := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root, "中").ID
	leaf := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &mid, "葉").ID
	depth := []struct {
		name string
		page string
	}{{"根", root}, {"中", mid}, {"葉", leaf}}

	alice := f.principalFor(ctx, t, f.alice)

	// 経路のどの段に、どの役割を張っても弱くならないことを見る。
	// 弱い役割を深い段へ張る組み合わせ（根 admin → 葉 viewer）が、降格が起きないことの肝。
	for _, tc := range []struct {
		name  string
		grant map[string]domain.GrantRole
	}{
		{"何も張らない", nil},
		{"根に viewer", map[string]domain.GrantRole{root: domain.GrantRoleViewer}},
		{"根に admin・葉に viewer", map[string]domain.GrantRole{
			root: domain.GrantRoleAdmin, leaf: domain.GrantRoleViewer,
		}},
		{"根に viewer・中に editor", map[string]domain.GrantRole{
			root: domain.GrantRoleViewer, mid: domain.GrantRoleEditor,
		}},
		{"中に admin だけ", map[string]domain.GrantRole{mid: domain.GrantRoleAdmin}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// 前の組み合わせを持ち越さない（付与は足し算なので、消さないと強い方が残る）。
			for _, d := range depth {
				require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, d.page, alice.ID))
			}
			for page, role := range tc.grant {
				f.grantPage(ctx, t, page, alice.ID, role)
			}

			// 1 ページの解決。
			var prev domain.PagePermission
			for i, d := range depth {
				got := f.permFor(ctx, t, d.page, f.alice)
				if i > 0 {
					assert.GreaterOrEqual(t, permRank(got), permRank(prev),
						"%s で %s より弱くなった（%+v → %+v）", d.name, depth[i-1].name, prev, got)
				}
				prev = got
			}

			// 一覧の解決も同じでなければならない。片方だけ単調でも、
			// 「開くと編集できるのに一覧では読み取り専用」というずれ方をする。
			rows, err := f.perm.ListSpacePageViewFacts(ctx, f.ws, f.spaceA, f.alice, false)
			require.NoError(t, err)
			view := map[string]bool{}
			for _, row := range rows {
				view[row.Page.ID] = domain.ResolvePageView(row.Role)
			}
			for i, d := range depth {
				if i > 0 && view[depth[i-1].page] {
					assert.True(t, view[d.page],
						"一覧: %s は見えるのに %s が見えない", depth[i-1].name, d.name)
				}
			}
		})
	}
}

// permRank は実効権限の強さを 1 つの整数に畳む（比較のためだけに使う）。
// できることが増えるほど大きくなる。
func permRank(p domain.PagePermission) int {
	n := 0
	if p.CanView {
		n++
	}
	if p.CanEdit {
		n++
	}
	if p.CanManage {
		n++
	}
	return n
}
