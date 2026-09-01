//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ページ単位の付与（page_grants）が、既定の役割としてどう効くかを実 PostgreSQL で固定する。
//
// 付与は「足す」層で、例外（page_restrictions）は「絞る」層。この 2 つの住み分けと、
// 経路（自分と祖先）をどう辿るかが、ここで確かめたいこと。
//
// **1 ページの解決と一覧の解決が割れないこと**を特に見る。既定の役割を計算する箇所が
// 6 つに分かれており、1 つでも枝を足し忘れると「開くと編集できるのに一覧では読み取り
// 専用に見える」というずれになる。経路ごとに別の答えを返すのが、この実装で最も起こり
// やすい壊れ方。

// grantPage はページに既定の役割を 1 行張る（書き込み API はまだ無いので直に入れる）。
func grantPage(t *testing.T, db *sql.DB, workspaceID, pageID, principalID string, role domain.GrantRole) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO page_grants (workspace_id, page_id, principal_id, "role")
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (workspace_id, page_id, principal_id)
		 DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()`,
		workspaceID, pageID, principalID, string(role),
	)
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

	grantPage(t, sqlDB, f.ws, page, alice.ID, domain.GrantRoleEditor)

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
	grantPage(t, sqlDB, f.ws, parent, alice.ID, domain.GrantRoleEditor)

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

	grantPage(t, sqlDB, f.ws, child, alice.ID, domain.GrantRoleEditor)

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
	grantPage(t, sqlDB, f.ws, page, alice.ID, domain.GrantRoleEditor)

	got := f.permFor(ctx, t, page, f.alice)
	assert.True(t, got.CanEdit, "ページの editor が採られる")

	// 逆向き（スペースが強い）でも、弱い付与が降格させないことを見る。
	f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleAdmin)
	grantPage(t, sqlDB, f.ws, page, alice.ID, domain.GrantRoleViewer)
	assert.True(t, f.permFor(ctx, t, page, f.alice).CanEdit,
		"ページに viewer を張ってもスペースの admin は下がらない（付与は足し算だけ）")
}

func TestPageGrant_例外は付与に勝つ_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象").ID
	alice := f.principalFor(ctx, t, f.alice)

	grantPage(t, sqlDB, f.ws, page, alice.ID, domain.GrantRoleAdmin)
	require.True(t, f.permFor(ctx, t, page, f.alice).CanView, "前提: 付与で見えている")

	// 同じページに deny を張る。例外は既定より強い。
	f.restrict(ctx, t, page, alice.ID, domain.CapabilityView, domain.RestrictionModeDeny)

	got := f.permFor(ctx, t, page, f.alice)
	assert.False(t, got.CanView, "deny は付与に勝つ（弱める操作は例外の層が担う）")
	assert.False(t, got.CanEdit)
}

func TestPageGrant_一覧と1ページの解決が一致する_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()
	f := setupKBPermission(t, sqlDB)

	parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
	child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
	outside := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象外").ID
	alice := f.principalFor(ctx, t, f.alice)

	grantPage(t, sqlDB, f.ws, parent, alice.ID, domain.GrantRoleEditor)

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

	grantPage(t, sqlDB, f.ws, page, everyone.ID, domain.GrantRoleEditor)

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
	assert.True(t, domain.ResolvePageView(rows[0].Facts), "ID 指定でも見える")
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
