//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createUser は users に 1 行入れて id を返す。principals / share_links は users へ FK を持つため、
// 権限の結合テストは実在するユーザーを前提にする（骨格側のテストが created_by_user_id に
// 固定値を使えるのは、pages が users へ FK を持たないから）。
//
// users は kbTables に含めない（ほかの結合テストと共有するため消さない）。代わりに毎回
// 一意なアドレスで作る。users には有効なユーザーのメールを一意にする部分索引があり、
// 固定アドレスを使い回すとサブテストの 2 回目で衝突する。
//
// id はシーケンスに任せるが、その前に必ず現在の最大 id へ合わせ直す。同じパッケージには
// TRUNCATE ... RESTART IDENTITY のあとに id を明示指定して users を作るテストがあり、
// そちらが通ると行だけが進んでシーケンスは 1 のまま取り残される。ここで採番すると
// その明示 id にぶつかって users_pkey が重複する。どのテストと組んでも成り立つように、
// 実行順（-shuffle）に依存しない形で毎回そろえる。
func createUser(t *testing.T, db *sql.DB, namePrefix string) uint64 {
	t.Helper()
	_, err := db.Exec(
		`SELECT setval('users_id_seq', COALESCE((SELECT max(id) FROM users), 0) + 1, false)`,
	)
	require.NoError(t, err)
	var id uint64
	require.NoError(t, db.QueryRow(
		`INSERT INTO users (email, name, role_id, is_active, created_at, updated_at)
		 VALUES ($1, $2, 3, true, now(), now()) RETURNING id`,
		namePrefix+"+"+newID()+"@example.test", namePrefix,
	).Scan(&id))
	return id
}

// kbPermFixture は権限の結合テストで使う共通の下ごしらえ。
type kbPermFixture struct {
	db       *sql.DB
	perm     repository.KnowledgeBasePermissionRepository
	pages    repository.KnowledgeBaseRepository
	pageUC   kbUseCases
	ws       string
	otherWS  string
	spaceA   string
	spaceB   string
	otherSpc string
	alice    uint64
	bob      uint64
	carol    uint64
}

func setupKBPermission(t *testing.T, gormDB *gorm.DB, sqlDB *sql.DB) kbPermFixture {
	t.Helper()
	testsupport.TruncateAll(t, gormDB, kbTables...)
	f := kbPermFixture{
		db:     sqlDB,
		perm:   persistence.NewKnowledgeBasePermissionRepository(sqlDB),
		pages:  persistence.NewKnowledgeBaseRepository(sqlDB),
		pageUC: newKbUseCases(persistence.NewKnowledgeBaseRepository(sqlDB)),
	}
	f.ws = createWorkspace(t, sqlDB, "perm-main")
	f.otherWS = createWorkspace(t, sqlDB, "perm-other")
	f.spaceA = createSpace(t, sqlDB, f.ws, "aaa")
	f.spaceB = createSpace(t, sqlDB, f.ws, "bbb")
	f.otherSpc = createSpace(t, sqlDB, f.otherWS, "ccc")
	f.alice = createUser(t, sqlDB, "alice")
	f.bob = createUser(t, sqlDB, "bob")
	f.carol = createUser(t, sqlDB, "carol")
	return f
}

// perm は 1 ページの実効権限を解いて返す（事実の収集 → 規則の適用の 2 段をまとめた小道具）。
func (f kbPermFixture) permFor(t *testing.T, pageID string, userID uint64) domain.PagePermission {
	t.Helper()
	facts, err := f.perm.PagePermissionFactsForUser(context.Background(), f.ws, pageID, userID)
	require.NoError(t, err)
	return domain.ResolvePagePermission(*facts)
}

// principalFor はユーザーの主体を用意して返す（ワークスペースへの所属追加も兼ねる）。
func (f kbPermFixture) principalFor(t *testing.T, userID uint64) *domain.Principal {
	t.Helper()
	p, err := f.perm.EnsureUserPrincipal(context.Background(), f.ws, userID)
	require.NoError(t, err)
	return p
}

// everyoneOf はそのスペースの「全員」の主体を用意して返す。
func (f kbPermFixture) everyoneOf(t *testing.T, spaceID string) *domain.Principal {
	t.Helper()
	p, err := f.perm.EnsureSpaceEveryonePrincipal(context.Background(), f.ws, spaceID)
	require.NoError(t, err)
	return p
}

// grantSpace はスペースの既定の役割を張る。
func (f kbPermFixture) grantSpace(t *testing.T, spaceID, principalID string, role domain.GrantRole) {
	t.Helper()
	_, err := f.perm.UpsertSpaceGrant(context.Background(), f.ws, spaceID, principalID, role)
	require.NoError(t, err)
}

// restrict はページに例外を 1 行張る。
func (f kbPermFixture) restrict(
	t *testing.T, pageID, principalID string, c domain.Capability, mode domain.RestrictionMode,
) {
	t.Helper()
	_, err := f.perm.UpsertPageRestriction(context.Background(), f.ws, pageID, principalID, c, mode)
	require.NoError(t, err)
}

// viewablePageIDs はそのユーザーに見えるページの ID を一覧経路（1 クエリ）で返す。
// 1 ページずつの解決と答えが割れないことを確かめるのに使う。
func (f kbPermFixture) viewablePageIDs(t *testing.T, spaceID string, userID uint64) []string {
	t.Helper()
	pages, err := usecase.NewListViewablePagesUseCase(f.perm).Execute(context.Background(),
		usecase.ListViewablePagesInput{WorkspaceID: f.ws, SpaceID: spaceID, UserID: userID})
	require.NoError(t, err)
	return pageIDs(pages)
}

func TestKnowledgeBasePermission_Integration(t *testing.T) {
	gormDB := testsupport.OpenTestDB(t)
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("メンバー追加は冪等で所属判定に効く", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)

		member, err := f.perm.IsWorkspaceMember(ctx, f.ws, f.alice)
		require.NoError(t, err)
		assert.False(t, member, "principal が無ければ非メンバー")

		first, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		second, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID, "2 回呼んでも主体は 1 つ")

		member, err = f.perm.IsWorkspaceMember(ctx, f.ws, f.alice)
		require.NoError(t, err)
		assert.True(t, member)

		// 別ワークスペースの所属は独立している。
		member, err = f.perm.IsWorkspaceMember(ctx, f.otherWS, f.alice)
		require.NoError(t, err)
		assert.False(t, member, "同じユーザーでも別テナントでは非メンバー")
	})

	t.Run("別ワークスペースのprincipalにgrantを張れない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		// alice を「別ワークスペース」のメンバーにして、その主体 ID を本命ワークスペースで使う。
		foreign, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.alice)
		require.NoError(t, err)

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, foreign.ID, domain.GrantRoleAdmin)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_space_grants_principal")

		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, foreign.ID, domain.GrantRoleAdmin)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_workspace_grants_principal")

		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, page.ID, foreign.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_page_restrictions_principal")
	})

	t.Run("存在しないユーザーのprincipalは作れない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		_, err := f.perm.EnsureUserPrincipal(ctx, f.ws, 999999999)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principals_user")
	})

	t.Run("ユーザーを消すと権限も消える", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		principal, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, principal.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		_, err = f.db.Exec(`DELETE FROM users WHERE id = $1`, f.bob)
		require.NoError(t, err)

		grants, err := f.perm.ListSpaceGrants(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		assert.Empty(t, grants, "ユーザーが消えたら principal も grant も残らない（別人への引き継ぎを作らない）")
	})

	t.Run("グループの入れ子はDBが弾く", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		outer, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "外側")
		require.NoError(t, err)
		inner, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "内側")
		require.NoError(t, err)

		err = f.perm.AddGroupMember(ctx, f.ws, outer.ID, inner.ID)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principal_members_member")

		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		err = f.perm.AddGroupMember(ctx, f.ws, alice.ID, inner.ID)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principal_members_group")
	})

	t.Run("kindごとに使う列がCHECKで固定されている", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)

		// kind='user' なのに user_id が無い。
		_, err := f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind) VALUES (gen_random_uuid(), $1, 'user')`, f.ws,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_user_id")

		// kind='group' なのに user_id が入っている。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind, user_id, name)
			 VALUES (gen_random_uuid(), $1, 'group', $2, '開発')`, f.ws, f.alice,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_user_id")

		// kind='group' なのに名前が空。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind) VALUES (gen_random_uuid(), $1, 'group')`, f.ws,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_name")

		// kind='share_link' なのに対象ページが無い（どのページのリンクか分からない主体は作らせない）。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind) VALUES (gen_random_uuid(), $1, 'share_link')`, f.ws,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_page_id")

		// kind='space_all' なのに対象スペースが無い。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind) VALUES (gen_random_uuid(), $1, 'space_all')`, f.ws,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_space_id")

		// 既知でない kind。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind) VALUES (gen_random_uuid(), $1, 'robot')`, f.ws,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_kind")
	})

	t.Run("スペース全員のprincipalは別テナントのスペースを指せない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		_, err := f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind, space_id)
			 VALUES (gen_random_uuid(), $1, 'space_all', $2)`, f.ws, f.otherSpc,
		)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principals_space")
	})

	t.Run("既定はスペースのgrantで決まる", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		assert.False(t, f.permFor(t, page.ID, f.alice).CanView, "grant が無ければ見えない")

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		got := f.permFor(t, page.ID, f.alice)
		assert.True(t, got.CanView)
		assert.False(t, got.CanEdit, "viewer は編集できない")

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, page.ID, f.alice).CanEdit, "同じ主体の grant は 1 行のまま更新される")

		grants, err := f.perm.ListSpaceGrants(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		require.Len(t, grants, 1, "upsert なので行は増えない")

		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, f.ws, f.spaceA, alice.ID))
		assert.False(t, f.permFor(t, page.ID, f.alice).CanView, "剥がせば見えなくなる")
	})

	t.Run("ワークスペースのgrantは全スペースに効きスペースのgrantで降格しない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		pageA := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "A ルート")
		pageB := mustCreatePage(t, f.pageUC, f.ws, f.spaceB, nil, "B ルート")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, pageA.ID, f.alice).CanEdit, "スペース A に grant が無くても効く")
		assert.True(t, f.permFor(t, pageB.ID, f.alice).CanEdit, "スペース B にも効く")

		// スペース側で弱い役割を張っても、強い方（ワークスペースの admin）が採られる。
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, pageA.ID, f.alice).CanEdit,
			"スペースに viewer を張るだけでワークスペース管理者を締め出せてはいけない")

		require.NoError(t, f.perm.DeleteWorkspaceGrant(ctx, f.ws, alice.ID))
		got := f.permFor(t, pageA.ID, f.alice)
		assert.True(t, got.CanView, "スペースの viewer が残る")
		assert.False(t, got.CanEdit)
	})

	t.Run("グループ経由とスペース全員の権限が効く", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		_, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, page.ID, f.alice).CanView, "スペース全員の grant はメンバーに効く")
		assert.False(t, f.permFor(t, page.ID, f.carol).CanView, "非メンバーには効かない")

		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "開発")
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, group.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		bobPrincipal, err := f.perm.FindUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, bobPrincipal.ID))

		assert.True(t, f.permFor(t, page.ID, f.bob).CanEdit, "グループ経由で editor")
		assert.False(t, f.permFor(t, page.ID, f.alice).CanEdit, "グループに入っていない人は viewer のまま")

		require.NoError(t, f.perm.RemoveGroupMember(ctx, f.ws, group.ID, bobPrincipal.ID))
		assert.False(t, f.permFor(t, page.ID, f.bob).CanEdit, "外せば既定に戻る")
	})

	t.Run("祖先の例外が効き近い段が優先され解除で既定に戻る", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")
		grand := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &child.ID, "grand")

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, grand.ID, f.alice).CanView, "例外が無ければ既定どおり")

		// child に「alice を deny」。孫まで継承される。
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, child.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeDeny)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, root.ID, f.alice).CanView, "祖先側（root）は影響を受けない")
		assert.False(t, f.permFor(t, child.ID, f.alice).CanView)
		assert.False(t, f.permFor(t, grand.ID, f.alice).CanView, "子孫へ継承される")
		assert.True(t, f.permFor(t, grand.ID, f.bob).CanView, "deny だけの段は他人の既定を変えない")

		// grand 自身に「alice を allow」を足しても、祖先の deny は覆らない。
		// deny を近い allow で外せると、自分の子ページを 1 枚作ってそこに自分への allow を
		// 張るだけで祖先の除外から逃げられてしまう。
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, grand.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		require.NoError(t, err)
		assert.False(t, f.permFor(t, grand.ID, f.alice).CanView, "経路上の deny は近い段の allow に勝つ")
		assert.False(t, f.permFor(t, grand.ID, f.bob).CanView, "allow リストに載っていない人は締め出される")

		// grand の allow を消すと、bob は許可リストから解放されて既定（editor）へ戻る。
		require.NoError(t, f.perm.DeletePageRestriction(ctx, f.ws, grand.ID, alice.ID, domain.CapabilityView))
		assert.False(t, f.permFor(t, grand.ID, f.alice).CanView, "child の deny は残っている")
		assert.True(t, f.permFor(t, grand.ID, f.bob).CanView)

		// すべて消すと既定（editor）に戻る。
		require.NoError(t, f.perm.DeletePageRestriction(ctx, f.ws, child.ID, alice.ID, domain.CapabilityView))
		assert.True(t, f.permFor(t, grand.ID, f.alice).CanEdit, "例外が無くなれば既定へ戻る")
	})

	t.Run("より近い許可リストがより遠い許可リストを上書きする", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")
		grand := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &child.ID, "grand")

		f.grantSpace(t, f.spaceA, f.everyoneOf(t, f.spaceA).ID, domain.GrantRoleEditor)
		alice := f.principalFor(t, f.alice)
		bob := f.principalFor(t, f.bob)

		// root は alice だけの限定公開。child でその枝だけ bob にも広げる。
		f.restrict(t, root.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		f.restrict(t, child.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		f.restrict(t, child.ID, bob.ID, domain.CapabilityView, domain.RestrictionModeAllow)

		assert.False(t, f.permFor(t, root.ID, f.bob).CanView, "root では bob は許可リストに載っていない")
		assert.True(t, f.permFor(t, child.ID, f.bob).CanView, "child 以下だけ広げられる")
		assert.True(t, f.permFor(t, grand.ID, f.bob).CanView, "child の許可リストは子孫にも効く")
		assert.ElementsMatch(t, []string{child.ID, grand.ID}, f.viewablePageIDs(t, f.spaceA, f.bob))
		assert.ElementsMatch(t, []string{root.ID, child.ID, grand.ID}, f.viewablePageIDs(t, f.spaceA, f.alice))
	})

	t.Run("祖先の限定公開は子孫のdeny1行で解除されない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		parent := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "人事・機密")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &parent.ID, "査定シート")
		grand := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &child.ID, "評価コメント")

		alice := f.principalFor(t, f.alice)
		bob := f.principalFor(t, f.bob)
		carol := f.principalFor(t, f.carol)
		for _, p := range []*domain.Principal{alice, bob, carol} {
			f.grantSpace(t, f.spaceA, p.ID, domain.GrantRoleEditor)
		}

		// parent を alice だけの限定公開にする（閲覧も編集も）。
		f.restrict(t, parent.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		f.restrict(t, parent.ID, alice.ID, domain.CapabilityEdit, domain.RestrictionModeAllow)
		for _, page := range []*domain.Page{parent, child, grand} {
			got := f.permFor(t, page.ID, f.bob)
			require.False(t, got.CanView, "限定公開の時点で bob には見えない")
			require.False(t, got.CanEdit)
		}
		require.Empty(t, f.viewablePageIDs(t, f.spaceA, f.bob))

		// 「carol だけ外す」という通常運用の deny を子ページに 1 行ずつ足す。
		f.restrict(t, child.ID, carol.ID, domain.CapabilityView, domain.RestrictionModeDeny)
		f.restrict(t, child.ID, carol.ID, domain.CapabilityEdit, domain.RestrictionModeDeny)

		for _, page := range []*domain.Page{child, grand} {
			got := f.permFor(t, page.ID, f.bob)
			assert.False(t, got.CanView, "第三者への deny 1 行で祖先の限定公開が解除されてはいけない")
			assert.False(t, got.CanEdit, "読めないページを編集できてもいけない")
		}
		assert.Empty(t, f.viewablePageIDs(t, f.spaceA, f.bob), "ツリー一覧にも露出しない")

		// deny は名指しされた本人にだけ効き、許可された本人の権限は変わらない。
		aliceOnChild := f.permFor(t, child.ID, f.alice)
		assert.True(t, aliceOnChild.CanView, "許可リストに載っている本人はそのまま")
		assert.True(t, aliceOnChild.CanEdit)
		assert.False(t, f.permFor(t, child.ID, f.carol).CanView, "名指しで外された本人は見えない")
		assert.ElementsMatch(t, []string{parent.ID, child.ID, grand.ID},
			f.viewablePageIDs(t, f.spaceA, f.alice), "限定公開は子孫まで効き続ける")
	})

	t.Run("読み取り専用のサブツリーは子のedit_deny1行で崩れない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "規程集")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "就業規則")

		f.grantSpace(t, f.spaceA, f.everyoneOf(t, f.spaceA).ID, domain.GrantRoleEditor)
		alice := f.principalFor(t, f.alice)
		f.principalFor(t, f.bob)
		carol := f.principalFor(t, f.carol)

		// root 以下は「alice だけが編集できる」読み取り専用サブツリー。
		f.restrict(t, root.ID, alice.ID, domain.CapabilityEdit, domain.RestrictionModeAllow)
		require.True(t, f.permFor(t, child.ID, f.bob).CanView, "閲覧の既定は editor のまま")
		require.False(t, f.permFor(t, child.ID, f.bob).CanEdit)

		// 「carol にだけは触らせない」つもりの deny を子に 1 行足す。
		f.restrict(t, child.ID, carol.ID, domain.CapabilityEdit, domain.RestrictionModeDeny)

		assert.False(t, f.permFor(t, child.ID, f.bob).CanEdit,
			"読み取り専用サブツリーが editor 全員に開いてはいけない（データ破壊になる）")
		assert.True(t, f.permFor(t, child.ID, f.alice).CanEdit, "許可された本人は編集できるまま")
		assert.False(t, f.permFor(t, child.ID, f.carol).CanEdit, "名指しで外された本人は編集できない")
	})

	// 次の 2 件は「閲覧の最も近い許可リストの段」と「編集のそれ」が別の depth にあり、
	// かつ片方の段の depth に、もう片方で名指しされた人の行が居る配置。
	// ケイパビリティごとの突き合わせを外すと、その人が相手側の段に載っているものとして
	// 通ってしまう。深さだけが一致していて意味は別、という取り違えを固定する。
	t.Run("閲覧の許可リストの段は編集の段と取り違えられない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")

		f.grantSpace(t, f.spaceA, f.everyoneOf(t, f.spaceA).ID, domain.GrantRoleEditor)
		alice := f.principalFor(t, f.alice)
		bob := f.principalFor(t, f.bob)
		carol := f.principalFor(t, f.carol)

		// child から見た深さは child=0 / root=1。
		// 閲覧の最も近い許可リストは child（bob だけ）で、alice の閲覧 allow は root にある。
		// 編集の最も近い許可リストは root（carol だけ）＝ alice の閲覧 allow と同じ深さ。
		f.restrict(t, root.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		f.restrict(t, child.ID, bob.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		f.restrict(t, root.ID, carol.ID, domain.CapabilityEdit, domain.RestrictionModeAllow)

		assert.False(t, f.permFor(t, child.ID, f.alice).CanView,
			"alice の閲覧 allow は child の許可リストより遠い段にある")
		assert.True(t, f.permFor(t, child.ID, f.bob).CanView, "最も近い許可リストに載っている本人は見える")
		assert.True(t, f.permFor(t, root.ID, f.alice).CanView, "root では alice が許可リストに載っている")
	})

	t.Run("編集の許可リストの段は閲覧の段と取り違えられない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")

		f.grantSpace(t, f.spaceA, f.everyoneOf(t, f.spaceA).ID, domain.GrantRoleEditor)
		alice := f.principalFor(t, f.alice)
		bob := f.principalFor(t, f.bob)

		// 編集の最も近い許可リストは child（bob だけ）で、alice の編集 allow は root にある。
		// 閲覧の最も近い許可リストは root（alice）＝ alice の編集 allow と同じ深さ。
		f.restrict(t, root.ID, alice.ID, domain.CapabilityEdit, domain.RestrictionModeAllow)
		f.restrict(t, child.ID, bob.ID, domain.CapabilityEdit, domain.RestrictionModeAllow)
		f.restrict(t, root.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)

		got := f.permFor(t, child.ID, f.alice)
		assert.True(t, got.CanView, "閲覧は root の許可リストで決まり alice は載っている")
		assert.False(t, got.CanEdit, "alice の編集 allow は child の許可リストより遠い段にある")
		assert.False(t, f.permFor(t, child.ID, f.bob).CanView, "bob は閲覧の許可リストに載っていない")
		assert.False(t, f.permFor(t, child.ID, f.bob).CanEdit, "閲覧できないので編集もできない")
	})

	t.Run("一覧は別ワークスペースのスペースとアーカイブ済みを返さない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		leaving := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "アーカイブする子")
		f.grantSpace(t, f.spaceA, f.principalFor(t, f.alice).ID, domain.GrantRoleEditor)
		require.ElementsMatch(t, []string{root.ID, leaving.ID}, f.viewablePageIDs(t, f.spaceA, f.alice))

		require.NoError(t, f.pageUC.archive.Execute(ctx, usecase.ArchivePageInput{
			WorkspaceID: f.ws, PageID: leaving.ID,
		}))
		assert.ElementsMatch(t, []string{root.ID}, f.viewablePageIDs(t, f.spaceA, f.alice),
			"アーカイブ済みは一覧に出ない")

		// 別ワークスペースのスペース ID を渡しても 1 枚も返さない（事実の収集の時点で塞ぐ）。
		mustCreatePage(t, f.pageUC, f.otherWS, f.otherSpc, nil, "別テナントのページ")
		foreign, err := f.perm.ListSpacePageViewFacts(ctx, f.ws, f.otherSpc, f.alice)
		require.NoError(t, err)
		assert.Empty(t, foreign, "テナント越えの spaceID では 0 件")
	})

	t.Run("スペース全員のdenyは別スペースへの移動で失効しない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		parent := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "親")
		secret := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &parent.ID, "社外秘")

		// 移動元・移動先とも全員 editor（移動先でも既定は変わらない）。
		for _, space := range []string{f.spaceA, f.spaceB} {
			f.grantSpace(t, space, f.everyoneOf(t, space).ID, domain.GrantRoleEditor)
		}
		f.principalFor(t, f.bob)
		everyoneA := f.everyoneOf(t, f.spaceA)
		f.restrict(t, secret.ID, everyoneA.ID, domain.CapabilityView, domain.RestrictionModeDeny)
		require.False(t, f.permFor(t, secret.ID, f.bob).CanView)

		// 例外を持つページの祖先を、別スペースのルートへ動かす（正規の操作）。
		_, err := f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.Error(t, err, "実効権限が緩む移動は失敗させる")
		assert.False(t, f.permFor(t, secret.ID, f.bob).CanView, "移動していないので deny は効いたまま")

		require.ErrorIs(t, err, repository.ErrPageMoveVoidsSpaceRestriction)
		moved, err := f.pages.FindPage(ctx, f.ws, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, f.spaceA, moved.SpaceID, "失敗した移動はロールバックされている")

		// allow でも同じく止める（deny だけフェイルオープン、という非対称を残さない）。
		require.NoError(t, f.perm.DeletePageRestriction(ctx, f.ws, secret.ID, everyoneA.ID, domain.CapabilityView))
		f.restrict(t, secret.ID, everyoneA.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.ErrorIs(t, err, repository.ErrPageMoveVoidsSpaceRestriction, "allow も同じ扱い")

		// 例外を先に整理すれば移せる（止めるのは「意味を失う例外が残っているとき」だけ）。
		require.NoError(t, f.perm.DeletePageRestriction(ctx, f.ws, secret.ID, everyoneA.ID, domain.CapabilityView))
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.NoError(t, err)
		moved, err = f.pages.FindPage(ctx, f.ws, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, f.spaceB, moved.SpaceID)
	})

	t.Run("移動先スペース全員の例外と同一スペース内の移動は止めない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)

		// スペースが変わらない移動は、例外の意味も変わらないので止めない。
		staying := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "親")
		newParent := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "別の親")
		stayingChild := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &staying.ID, "社外秘")
		f.restrict(t, stayingChild.ID, f.everyoneOf(t, f.spaceA).ID,
			domain.CapabilityView, domain.RestrictionModeDeny)
		_, err := f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: staying.ID, NewParentID: &newParent.ID,
		})
		require.NoError(t, err, "同一スペース内の移動は例外の意味を変えない")

		// 「移動先スペースの全員」宛ての例外は、移動後にこそ意味を持つので止めない。
		leaving := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "移すページ")
		leavingChild := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &leaving.ID, "移す子")
		f.restrict(t, leavingChild.ID, f.everyoneOf(t, f.spaceB).ID,
			domain.CapabilityView, domain.RestrictionModeDeny)
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: leaving.ID, NewSpaceID: f.spaceB,
		})
		require.NoError(t, err, "移動先スペース宛ての例外は移動後に効くので止めない")
	})

	t.Run("同じ主体とケイパビリティに矛盾する例外は作れない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, page.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		require.NoError(t, err)
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, page.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeDeny)
		require.NoError(t, err)

		rows, err := f.perm.ListPageRestrictions(ctx, f.ws, page.ID)
		require.NoError(t, err)
		require.Len(t, rows, 1, "allow と deny の 2 行にはならない（PK が向きを 1 つに絞る）")
		assert.Equal(t, domain.RestrictionModeDeny, rows[0].Mode)
	})

	t.Run("祖先をたどる解決が別スペースへ漏れない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		// スペース A と B に、それぞれ独立した木を作る。
		aRoot := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "A ルート")
		aChild := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &aRoot.ID, "A 子")
		bRoot := mustCreatePage(t, f.pageUC, f.ws, f.spaceB, nil, "B ルート")

		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		for _, space := range []string{f.spaceA, f.spaceB} {
			everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, space)
			require.NoError(t, err)
			_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, space, everyone.ID, domain.GrantRoleEditor)
			require.NoError(t, err)
		}

		// A ルートを限定公開（alice だけ）にしても、B の木には影響しない。
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, aRoot.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		require.NoError(t, err)
		assert.True(t, f.permFor(t, aChild.ID, f.alice).CanView)
		assert.False(t, f.permFor(t, aChild.ID, f.bob).CanView, "A の木は限定公開になった")
		assert.True(t, f.permFor(t, bRoot.ID, f.bob).CanView, "B の木は無関係のまま")

		// スペース全員の grant もスペースごとに独立している。
		everyoneA, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, f.ws, f.spaceA, everyoneA.ID))
		assert.False(t, f.permFor(t, aRoot.ID, f.bob).CanView)
		assert.True(t, f.permFor(t, bRoot.ID, f.bob).CanView, "A の grant を剥がしても B は変わらない")
	})

	t.Run("ページを動かすと実効権限が変わる", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		open := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "公開の親")
		secret := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "秘密の親")
		moving := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &open.ID, "動くページ")

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		// 「秘密の親」以下は alice だけ。
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, secret.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		require.NoError(t, err)

		assert.True(t, f.permFor(t, moving.ID, f.bob).CanView, "公開の親の下では bob も見える")

		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: moving.ID, NewParentID: &secret.ID,
		})
		require.NoError(t, err)
		assert.False(t, f.permFor(t, moving.ID, f.bob).CanView, "秘密の親の下へ移すと見えなくなる")
		assert.True(t, f.permFor(t, moving.ID, f.alice).CanView)

		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: moving.ID, NewParentID: &open.ID,
		})
		require.NoError(t, err)
		assert.True(t, f.permFor(t, moving.ID, f.bob).CanView, "戻せばまた見える（例外の行は 1 つも触っていない）")
	})

	t.Run("閲覧可能ページ一覧は見えないページを落とす", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		open := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "公開")
		secret := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "秘密")
		secretChild := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &secret.ID, "秘密の子")

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, secret.ID, alice.ID, domain.CapabilityView, domain.RestrictionModeAllow)
		require.NoError(t, err)

		listUC := usecase.NewListViewablePagesUseCase(f.perm)
		bobPages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.bob,
		})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{root.ID, open.ID}, pageIDs(bobPages), "秘密の木は丸ごと落ちる")

		alicePages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.alice,
		})
		require.NoError(t, err)
		assert.ElementsMatch(t,
			[]string{root.ID, open.ID, secret.ID, secretChild.ID}, pageIDs(alicePages),
			"許可された人には子孫まで見える")

		carolPages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.carol,
		})
		require.NoError(t, err)
		assert.Empty(t, carolPages, "非メンバーには 1 枚も見えない")
	})

	t.Run("メンバー削除のusecaseは主体ごと消す", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		require.True(t, f.permFor(t, page.ID, f.alice).CanEdit)

		removeUC := usecase.NewRemoveWorkspaceMemberUseCase(f.perm)
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}))
		assert.False(t, f.permFor(t, page.ID, f.alice).CanView, "所属を外すと権限も消える")
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}),
			"二度目は冪等に成功する")

		require.ErrorIs(t, f.perm.DeletePrincipal(ctx, f.ws, alice.ID), repository.ErrPrincipalNotFound)
	})

	t.Run("grantと例外の一覧を引ける", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		grantWS := usecase.NewGrantWorkspaceRoleUseCase(f.perm)
		_, err = grantWS.Execute(ctx, usecase.GrantWorkspaceRoleInput{
			WorkspaceID: f.ws, PrincipalID: alice.ID, Role: domain.GrantRoleAdmin,
		})
		require.NoError(t, err)
		wsGrants, err := f.perm.ListWorkspaceGrants(ctx, f.ws)
		require.NoError(t, err)
		require.Len(t, wsGrants, 1)
		assert.Equal(t, domain.GrantRoleAdmin, wsGrants[0].Role)

		setUC := usecase.NewSetPageRestrictionUseCase(f.perm)
		_, err = setUC.Execute(ctx, usecase.SetPageRestrictionInput{
			WorkspaceID: f.ws, PageID: page.ID, PrincipalID: alice.ID,
			Capability: domain.CapabilityEdit, Mode: domain.RestrictionModeDeny,
		})
		require.NoError(t, err)
		restrictions, err := f.perm.ListPageRestrictions(ctx, f.ws, page.ID)
		require.NoError(t, err)
		require.Len(t, restrictions, 1)
		assert.False(t, f.permFor(t, page.ID, f.alice).CanEdit, "admin でも例外の deny には勝てない")
		assert.True(t, f.permFor(t, page.ID, f.alice).CanView)

		clearUC := usecase.NewClearPageRestrictionUseCase(f.perm)
		require.NoError(t, clearUC.Execute(ctx, usecase.ClearPageRestrictionInput{
			WorkspaceID: f.ws, PageID: page.ID, PrincipalID: alice.ID, Capability: domain.CapabilityEdit,
		}))
		assert.True(t, f.permFor(t, page.ID, f.alice).CanEdit, "解除すれば既定へ戻る")

		require.NoError(t, usecase.NewRevokeWorkspaceRoleUseCase(f.perm).Execute(ctx,
			usecase.RevokeWorkspaceRoleInput{WorkspaceID: f.ws, PrincipalID: alice.ID}))
		assert.False(t, f.permFor(t, page.ID, f.alice).CanView)
	})

	t.Run("グループ操作のusecaseが権限に効く", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		_, err := usecase.NewAddWorkspaceMemberUseCase(f.perm).Execute(ctx,
			usecase.AddWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.bob})
		require.NoError(t, err)
		group, err := usecase.NewCreatePrincipalGroupUseCase(f.perm).Execute(ctx,
			usecase.CreatePrincipalGroupInput{WorkspaceID: f.ws, Name: "開発"})
		require.NoError(t, err)
		_, err = usecase.NewGrantSpaceRoleUseCase(f.perm).Execute(ctx, usecase.GrantSpaceRoleInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, PrincipalID: group.ID, Role: domain.GrantRoleEditor,
		})
		require.NoError(t, err)

		addUC := usecase.NewAddGroupMemberUseCase(f.perm)
		require.NoError(t, addUC.Execute(ctx, usecase.AddGroupMemberInput{
			WorkspaceID: f.ws, GroupPrincipalID: group.ID, MemberUserID: f.bob,
		}))
		require.NoError(t, addUC.Execute(ctx, usecase.AddGroupMemberInput{
			WorkspaceID: f.ws, GroupPrincipalID: group.ID, MemberUserID: f.bob,
		}), "同じ人を二度加えても冪等")
		assert.True(t, f.permFor(t, page.ID, f.bob).CanEdit)

		removeUC := usecase.NewRemoveGroupMemberUseCase(f.perm)
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveGroupMemberInput{
			WorkspaceID: f.ws, GroupPrincipalID: group.ID, MemberUserID: f.bob,
		}))
		assert.False(t, f.permFor(t, page.ID, f.bob).CanView)

		// スペース全員の主体も usecase 経由で用意でき、二度呼んでも増えない。
		everyoneUC := usecase.NewEnsureSpaceEveryonePrincipalUseCase(f.perm)
		first, err := everyoneUC.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{WorkspaceID: f.ws, SpaceID: f.spaceA})
		require.NoError(t, err)
		second, err := everyoneUC.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{WorkspaceID: f.ws, SpaceID: f.spaceA})
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID)
	})

	t.Run("形式が不正なIDは存在しないものとして扱う", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		const bad = "not-a-uuid"

		_, err := f.perm.FindPrincipal(ctx, bad, bad)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
		_, err = f.perm.FindUserPrincipal(ctx, bad, f.alice)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
		require.ErrorIs(t, f.perm.DeletePrincipal(ctx, bad, bad), repository.ErrPrincipalNotFound)
		require.ErrorIs(t, f.perm.RevokeShareLink(ctx, bad, bad), repository.ErrShareLinkNotFound)
		_, err = f.perm.PagePermissionFactsForUser(ctx, bad, bad, f.alice)
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = f.perm.PagePermissionFactsForPrincipal(ctx, f.ws, f.spaceA, bad)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)

		member, err := f.perm.IsWorkspaceMember(ctx, bad, f.alice)
		require.NoError(t, err)
		assert.False(t, member)

		// 一覧系は空を返す（URL 由来の生文字列を DB エラーにしない）。
		wsGrants, err := f.perm.ListWorkspaceGrants(ctx, bad)
		require.NoError(t, err)
		assert.Empty(t, wsGrants)
		spGrants, err := f.perm.ListSpaceGrants(ctx, bad, bad)
		require.NoError(t, err)
		assert.Empty(t, spGrants)
		restrictions, err := f.perm.ListPageRestrictions(ctx, bad, bad)
		require.NoError(t, err)
		assert.Empty(t, restrictions)
		links, err := f.perm.ListPageShareLinks(ctx, bad, bad)
		require.NoError(t, err)
		assert.Empty(t, links)
		facts, err := f.perm.ListSpacePageViewFacts(ctx, bad, bad, f.alice)
		require.NoError(t, err)
		assert.Empty(t, facts)

		require.NoError(t, f.perm.RemoveGroupMember(ctx, bad, bad, bad))
		require.NoError(t, f.perm.DeleteWorkspaceGrant(ctx, bad, bad))
		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, bad, bad, bad))
		require.NoError(t, f.perm.DeletePageRestriction(ctx, bad, bad, bad, domain.CapabilityView))
		require.ErrorIs(t, f.perm.AddGroupMember(ctx, bad, bad, bad), repository.ErrPrincipalNotFound)
		_, err = f.perm.EnsureUserPrincipal(ctx, bad, f.alice)
		require.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
		_, err = f.perm.EnsureSpaceEveryonePrincipal(ctx, bad, bad)
		require.ErrorIs(t, err, repository.ErrSpaceNotFound)
		_, err = f.perm.CreateGroupPrincipal(ctx, bad, "x")
		require.ErrorIs(t, err, repository.ErrWorkspaceNotFound)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, bad, bad, domain.GrantRoleViewer)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
		_, err = f.perm.UpsertSpaceGrant(ctx, bad, bad, bad, domain.GrantRoleViewer)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
		_, err = f.perm.UpsertPageRestriction(ctx, bad, bad, bad, domain.CapabilityView, domain.RestrictionModeAllow)
		require.ErrorIs(t, err, repository.ErrPageNotFound)
		_, err = f.perm.CreateShareLink(ctx, repository.ShareLinkWrite{WorkspaceID: bad, PageID: bad})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
	})

	t.Run("別テナントの所属は解決に持ち込まれない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		// alice は「別ワークスペース」だけのメンバーで、そちらでは admin。
		foreign, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.alice)
		require.NoError(t, err)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.otherWS, foreign.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)
		// こちらのスペースは全員 editor。
		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		facts, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, page.ID, f.alice)
		require.NoError(t, err)
		assert.False(t, facts.Member, "別テナントの主体をこちらの所属として拾ってはいけない")
		assert.Nil(t, facts.Role, "別テナントの grant を持ち込んではいけない")
		assert.False(t, domain.ResolvePagePermission(*facts).CanView,
			"スペース全員の grant は非メンバーには効かない")
	})

	t.Run("別ワークスペースのページは解決できない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		foreignPage := mustCreatePage(t, f.pageUC, f.otherWS, f.otherSpc, nil, "別テナントのページ")

		_, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, foreignPage.ID, f.alice)
		require.ErrorIs(t, err, repository.ErrPageNotFound, "テナント越えは「無い」と同じ扱い")
	})
}

func TestKnowledgeBaseShareLink_Integration(t *testing.T) {
	gormDB := testsupport.OpenTestDB(t)
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)
	ctx := context.Background()

	t.Run("発行したリンクで対象ページと子孫を開ける", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		child := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "子")
		outside := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "対象外")

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		verified, err := usecase.NewVerifyShareLinkUseCase(f.perm).
			Execute(ctx, usecase.VerifyShareLinkInput{Token: issued.Token})
		require.NoError(t, err)

		checkUC := usecase.NewCheckShareLinkPermissionUseCase(f.perm, f.pages)
		got, err := checkUC.Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: verified, PageID: child.ID})
		require.NoError(t, err)
		assert.True(t, got.CanView)
		assert.False(t, got.CanEdit, "閲覧のリンクでは編集できない")

		_, err = checkUC.Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: verified, PageID: outside.ID})
		require.ErrorIs(t, err, usecase.ErrShareLinkPageOutOfScope, "リンクの木の外は開けない")
	})

	t.Run("平文トークンはDBに残らない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView,
			Password: "s3cret", CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		var count int
		require.NoError(t, f.db.QueryRow(
			`SELECT count(*) FROM share_links WHERE encode(token_hash, 'escape') = $1 OR password_hash = $2`,
			issued.Token, "s3cret",
		).Scan(&count))
		assert.Zero(t, count, "トークンもパスワードも平文では入っていない")

		var hashLen int
		require.NoError(t, f.db.QueryRow(
			`SELECT octet_length(token_hash) FROM share_links WHERE id = $1`, issued.Link.ID,
		).Scan(&hashLen))
		assert.Equal(t, 32, hashLen, "SHA-256 の 32 バイト")
	})

	t.Run("期限切れと失効は開けない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		issueUC := usecase.NewIssueShareLinkUseCase(f.perm)
		verifyUC := usecase.NewVerifyShareLinkUseCase(f.perm)

		expiring, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView,
			ExpiresAt: ptrTime(time.Now().Add(time.Hour)), CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		_, err = verifyUC.Execute(ctx, usecase.VerifyShareLinkInput{Token: expiring.Token})
		require.NoError(t, err, "期限内は開ける")

		// 期限を過去へ倒す（時間を待たずに期限切れを再現する）。
		_, err = f.db.Exec(`UPDATE share_links SET expires_at = now() - interval '1 minute' WHERE id = $1`, expiring.Link.ID)
		require.NoError(t, err)
		_, err = verifyUC.Execute(ctx, usecase.VerifyShareLinkInput{Token: expiring.Token})
		require.ErrorIs(t, err, usecase.ErrShareLinkExpired)

		revoking, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		revokeUC := usecase.NewRevokeShareLinkUseCase(f.perm)
		require.NoError(t, revokeUC.Execute(ctx, usecase.RevokeShareLinkInput{
			WorkspaceID: f.ws, ShareLinkID: revoking.Link.ID,
		}))
		_, err = verifyUC.Execute(ctx, usecase.VerifyShareLinkInput{Token: revoking.Token})
		require.ErrorIs(t, err, usecase.ErrShareLinkRevoked)

		require.NoError(t, revokeUC.Execute(ctx, usecase.RevokeShareLinkInput{
			WorkspaceID: f.ws, ShareLinkID: revoking.Link.ID,
		}), "二度目の失効は冪等に成功する")
		require.ErrorIs(t, revokeUC.Execute(ctx, usecase.RevokeShareLinkInput{
			WorkspaceID: f.otherWS, ShareLinkID: revoking.Link.ID,
		}), repository.ErrShareLinkNotFound, "別テナントからは失効させられない")
	})

	t.Run("公開しつつ子ページだけdenyで隠せる", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		hidden := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "隠す子")

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		_, err = f.perm.UpsertPageRestriction(ctx, f.ws, hidden.ID, issued.Link.PrincipalID,
			domain.CapabilityView, domain.RestrictionModeDeny)
		require.NoError(t, err)

		verified, err := usecase.NewVerifyShareLinkUseCase(f.perm).
			Execute(ctx, usecase.VerifyShareLinkInput{Token: issued.Token})
		require.NoError(t, err)
		checkUC := usecase.NewCheckShareLinkPermissionUseCase(f.perm, f.pages)

		rootPerm, err := checkUC.Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: verified, PageID: root.ID})
		require.NoError(t, err)
		assert.True(t, rootPerm.CanView)

		hiddenPerm, err := checkUC.Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: verified, PageID: hidden.ID})
		require.NoError(t, err)
		assert.False(t, hiddenPerm.CanView, "リンクの主体を deny した子ページは開けない")
	})

	t.Run("除外した子ページは無関係なdenyで復活しない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		hidden := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "隠す子")
		grand := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &hidden.ID, "隠す子の子")

		issueUC := usecase.NewIssueShareLinkUseCase(f.perm)
		viewLink, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		editLink, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityEdit, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		// どちらのリンクからも「隠す子」以下を除外する（deny 1 行）。
		f.restrict(t, hidden.ID, viewLink.Link.PrincipalID, domain.CapabilityView, domain.RestrictionModeDeny)
		f.restrict(t, hidden.ID, editLink.Link.PrincipalID, domain.CapabilityView, domain.RestrictionModeDeny)

		linkPerm := shareLinkPermFunc(t, f)
		require.False(t, linkPerm(viewLink.Token, hidden.ID).CanView)
		require.False(t, linkPerm(viewLink.Token, grand.ID).CanView, "除外はサブツリー全体に効く")

		// 除外した子の配下に、リンクとは無関係なメンバーの deny が 1 行付く。
		carol := f.principalFor(t, f.carol)
		f.restrict(t, grand.ID, carol.ID, domain.CapabilityView, domain.RestrictionModeDeny)

		assert.False(t, linkPerm(viewLink.Token, hidden.ID).CanView)
		assert.False(t, linkPerm(viewLink.Token, grand.ID).CanView,
			"無関係な deny 1 行で未認証の来訪者に露出してはいけない")
		editPerm := linkPerm(editLink.Token, grand.ID)
		assert.False(t, editPerm.CanView, "編集リンクでも同じ")
		assert.False(t, editPerm.CanEdit, "未認証で編集が通ってはいけない")
	})

	t.Run("一般メンバーの自己denyでは公開リンクに露出しない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		root := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		hidden := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &root.ID, "社外秘")
		grand := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, &hidden.ID, "社外秘の子")

		f.grantSpace(t, f.spaceA, f.everyoneOf(t, f.spaceA).ID, domain.GrantRoleEditor)
		bob := f.principalFor(t, f.bob)

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		f.restrict(t, hidden.ID, issued.Link.PrincipalID, domain.CapabilityView, domain.RestrictionModeDeny)
		linkPerm := shareLinkPermFunc(t, f)
		require.False(t, linkPerm(issued.Token, grand.ID).CanView)

		// 一般メンバーの bob が「自分自身の閲覧を deny」する 1 行を張るだけ。
		// 共有リンクにも権限設定画面にも触っていない。
		_, err = usecase.NewSetPageRestrictionUseCase(f.perm).Execute(ctx, usecase.SetPageRestrictionInput{
			WorkspaceID: f.ws, PageID: grand.ID, PrincipalID: bob.ID,
			Capability: domain.CapabilityView, Mode: domain.RestrictionModeDeny,
		})
		require.NoError(t, err)

		assert.False(t, linkPerm(issued.Token, grand.ID).CanView,
			"自己 deny 1 行で社外秘ページを公開 URL に載せられてはいけない")
	})

	t.Run("共有リンクの主体はスペースのgrantを拾わない", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		facts, err := f.perm.PagePermissionFactsForPrincipal(ctx, f.ws, page.ID, issued.Link.PrincipalID)
		require.NoError(t, err)
		assert.False(t, facts.Member, "リンクの来訪者はメンバーではない")
		assert.Nil(t, facts.Role, "スペース全員の grant（admin）を拾ってはいけない")
	})

	t.Run("リンクを消すと主体も消える", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		// ページを物理削除すると share_links も principal も CASCADE で消える。
		_, err = f.db.Exec(`DELETE FROM pages WHERE id = $1`, page.ID)
		require.NoError(t, err)

		var links, principals int
		require.NoError(t, f.db.QueryRow(`SELECT count(*) FROM share_links WHERE id = $1`, issued.Link.ID).Scan(&links))
		require.NoError(t, f.db.QueryRow(`SELECT count(*) FROM principals WHERE id = $1`, issued.Link.PrincipalID).Scan(&principals))
		assert.Zero(t, links)
		assert.Zero(t, principals, "リンクだけが消えて主体が残る、という状態を作らない")
	})

	t.Run("ページの共有リンク一覧は失効済みも返す", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		issueUC := usecase.NewIssueShareLinkUseCase(f.perm)
		alive, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		dead, err := issueUC.Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityEdit, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		require.NoError(t, usecase.NewRevokeShareLinkUseCase(f.perm).Execute(ctx,
			usecase.RevokeShareLinkInput{WorkspaceID: f.ws, ShareLinkID: dead.Link.ID}))

		links, err := f.perm.ListPageShareLinks(ctx, f.ws, page.ID)
		require.NoError(t, err)
		require.Len(t, links, 2, "失効済みも含めて返す（誰がいつ止めたかを追えるように）")
		byID := map[string]domain.ShareLink{}
		for _, l := range links {
			byID[l.ID] = l
		}
		assert.Nil(t, byID[alive.Link.ID].RevokedAt)
		assert.NotNil(t, byID[dead.Link.ID].RevokedAt)
	})

	t.Run("共有リンクの主体は対象ページに結び付く", func(t *testing.T) {
		f := setupKBPermission(t, gormDB, sqlDB)
		page := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "root")
		other := mustCreatePage(t, f.pageUC, f.ws, f.spaceA, nil, "別ページ")
		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		principal, err := f.perm.FindPrincipal(ctx, f.ws, issued.Link.PrincipalID)
		require.NoError(t, err)
		require.NotNil(t, principal.PageID)
		assert.Equal(t, page.ID, *principal.PageID)

		// リンクの page_id だけを別ページへ書き換えることはできない（主体と食い違わせない）。
		_, err = f.db.Exec(`UPDATE share_links SET page_id = $1 WHERE id = $2`, other.ID, issued.Link.ID)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_share_links_principal")

		// share_link 以外の主体は page_id を持てない。
		_, err = f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind, name, page_id)
			 VALUES (gen_random_uuid(), $1, 'group', '開発', $2)`, f.ws, page.ID,
		)
		requirePgError(t, err, sqlStateCheckViolation, "ck_principals_page_id")
	})
}

// shareLinkPermFunc は「このトークンの来訪者にこのページはどう見えるか」を解く小道具を返す。
// トークンの検証から権限解決までを毎回通すので、未認証の来訪者が実際に辿る経路と同じになる。
func shareLinkPermFunc(t *testing.T, f kbPermFixture) func(token, pageID string) domain.PagePermission {
	t.Helper()
	verifyUC := usecase.NewVerifyShareLinkUseCase(f.perm)
	checkUC := usecase.NewCheckShareLinkPermissionUseCase(f.perm, f.pages)
	return func(token, pageID string) domain.PagePermission {
		t.Helper()
		link, err := verifyUC.Execute(context.Background(), usecase.VerifyShareLinkInput{Token: token})
		require.NoError(t, err)
		got, err := checkUC.Execute(context.Background(),
			usecase.CheckShareLinkPermissionInput{Link: link, PageID: pageID})
		require.NoError(t, err)
		return *got
	}
}

// pageIDs はページの ID だけを取り出す（一覧の比較用）。
func pageIDs(pages []domain.Page) []string {
	ids := make([]string, 0, len(pages))
	for _, p := range pages {
		ids = append(ids, p.ID)
	}
	return ids
}

func ptrTime(t time.Time) *time.Time { return &t }
