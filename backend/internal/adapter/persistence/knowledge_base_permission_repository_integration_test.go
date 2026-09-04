//go:build integration

package persistence_test

import (
	"context"
	"database/sql"
	"slices"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		`INSERT INTO users (email, name, role, is_active, created_at, updated_at)
		 VALUES ($1, $2, 'trainee', true, now(), now()) RETURNING id`,
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

func setupKBPermission(t *testing.T, sqlDB *sql.DB) kbPermFixture {
	t.Helper()
	testsupport.TruncateAll(t, sqlDB, kbTables...)
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
func (f kbPermFixture) permFor(ctx context.Context, t *testing.T, pageID string, userID uint64) domain.PagePermission {
	t.Helper()
	facts, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, pageID, userID)
	require.NoError(t, err)
	return domain.ResolvePagePermission(*facts)
}

// principalFor はユーザーの主体を用意して返す（ワークスペースへの所属追加も兼ねる）。
func (f kbPermFixture) principalFor(ctx context.Context, t *testing.T, userID uint64) *domain.Principal {
	t.Helper()
	p, err := f.perm.EnsureUserPrincipal(ctx, f.ws, userID)
	require.NoError(t, err)
	return p
}

// everyoneOf はそのスペースの「全員」の主体を用意して返す。
func (f kbPermFixture) everyoneOf(ctx context.Context, t *testing.T, spaceID string) *domain.Principal {
	t.Helper()
	p, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, spaceID)
	require.NoError(t, err)
	return p
}

// grantSpace はスペースの既定の役割を張る。
func (f kbPermFixture) grantSpace(ctx context.Context, t *testing.T, spaceID, principalID string, role domain.GrantRole) {
	t.Helper()
	_, err := f.perm.UpsertSpaceGrant(ctx, f.ws, spaceID, principalID, role)
	require.NoError(t, err)
}

// makePrivate はスペースを private にする。
//
// **これが「見せない」を表す唯一のやり方**。権限は 3 段の付与（ワークスペース / スペース /
// ページ）を足し合わせ、届いた中で最も強い役割で決まるので、同じスペースの中で 1 枚だけ
// 隠すことはできない。private のスペースにはワークスペース全体の付与とスペース全員宛ての
// 付与が届かず、そのスペースを名指しした付与だけが届く。
//
// 生の UPDATE を使うのは、visibility を変える repository の口がまだ無いため
// （作成時に決める列で、テストだけがあとから倒したい）。
func (f kbPermFixture) makePrivate(t *testing.T, spaceID string) {
	t.Helper()
	res, err := f.db.Exec(
		`UPDATE spaces SET visibility = 'private', updated_at = now()
		 WHERE workspace_id = $1 AND id = $2`, f.ws, spaceID,
	)
	require.NoError(t, err)
	n, err := res.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), n, "対象のスペースが見つかりません")
}

// viewablePageIDs はそのユーザーに見えるページの ID を一覧経路（1 クエリ）で返す。
// 1 ページずつの解決と答えが割れないことを確かめるのに使う。
func (f kbPermFixture) viewablePageIDs(ctx context.Context, t *testing.T, spaceID string, userID uint64) []string {
	t.Helper()
	out, err := usecase.NewListViewablePagesUseCase(f.perm).Execute(ctx,
		usecase.ListViewablePagesInput{WorkspaceID: f.ws, SpaceID: spaceID, UserID: userID})
	require.NoError(t, err)
	return pageIDs(out.Pages)
}

func TestKnowledgeBaseSiblingPositionsAround_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("隣り合うキーを返し、端は空文字にする", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根")
		a := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "A")
		b := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "B")
		c := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "C")

		found, prev, at, next, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, b.ID, "")
		require.NoError(t, err)

		assert.True(t, found)
		assert.Equal(t, a.Position, prev)
		assert.Equal(t, b.Position, at)
		assert.Equal(t, c.Position, next)

		// 先頭は手前が空文字（fracindex.Between の「端」）。
		_, prevOfFirst, _, _, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, a.ID, "")
		require.NoError(t, err)
		assert.Empty(t, prevOfFirst)

		// 末尾は次が空文字。
		_, _, _, nextOfLast, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, c.ID, "")
		require.NoError(t, err)
		assert.Empty(t, nextOfLast)
	})

	t.Run("動かす当人は隣人に数えない", func(t *testing.T) {
		// 除かないと自分自身との中間値を計算することになる。
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根")
		a := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "A")
		b := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "B")
		c := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "C")

		// B を動かしながら A の隣を尋ねると、A の次は C になる（B は居ないものとして扱う）。
		_, _, _, next, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, a.ID, b.ID)
		require.NoError(t, err)
		assert.Equal(t, c.Position, next)

		// 自分自身を隣に指定したら「兄弟ではない」。
		found, _, _, _, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, b.ID, b.ID)
		require.NoError(t, err)
		assert.False(t, found)
	})

	t.Run("別の親・別スペース・アーカイブ済み・不在をまとめて「兄弟ではない」にする", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根")
		other := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "別の親")
		underOther := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &other.ID, "別の親の子")
		inSpaceB := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceB, nil, "別スペース")
		archived := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "アーカイブ")
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, archived.ID))

		for _, id := range []string{
			underOther.ID,
			inSpaceB.ID,
			archived.ID,
			"0198a000-0000-7000-8000-0000000000ff",
			"not-a-uuid",
		} {
			found, _, _, _, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, id, "")
			require.NoError(t, err, "id=%s", id)
			assert.False(t, found, "id=%s は root の現役の子ではない", id)
		}
	})

	t.Run("誰にも届いていない兄弟も並びには居るので隣人に数える", func(t *testing.T) {
		// 見えないだけで並びには居る。除くとキーが既存の行と衝突する。
		// このクエリは権限を一切見ない（見せるかどうかは呼び出し側の話）。
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根")
		a := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "A")
		unreachable := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "誰にも届いていない")
		// alice には A だけを届かせ、隣の 1 枚には何も張らない。
		alice := f.principalFor(ctx, t, f.alice)
		f.grantPage(ctx, t, a.ID, alice.ID, domain.GrantRoleEditor)
		require.False(t, f.permFor(ctx, t, unreachable.ID, f.alice).CanView, "前提: 隣は alice に見えない")

		_, _, _, next, err := f.pages.SiblingPositionsAround(ctx, f.ws, f.spaceA, &root.ID, a.ID, "")
		require.NoError(t, err)
		assert.Equal(t, unreachable.Position, next, "権限に関わらず並びの隣を返す")
	})
}

func TestKnowledgeBaseArchivedViewFacts_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	// listFor はその一覧に出るページ ID を返す（現役／アーカイブ済みを切り替える）。
	listFor := func(f kbPermFixture, t *testing.T, userID uint64, archived bool) []string {
		t.Helper()
		out, err := usecase.NewListViewablePagesUseCase(f.perm).Execute(ctx,
			usecase.ListViewablePagesInput{
				WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: userID, Archived: archived,
			})
		require.NoError(t, err)
		return pageIDs(out.Pages)
	}

	t.Run("現役とアーカイブ済みが混ざらない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		alive := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "現役")
		gone := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "アーカイブ")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, gone.ID))

		assert.Equal(t, []string{alive.ID}, listFor(f, t, f.alice, false))
		assert.Equal(t, []string{gone.ID}, listFor(f, t, f.alice, true))
	})

	t.Run("アーカイブ済みでも、届いていない相手には出ない", func(t *testing.T) {
		// **この検査が本命。** 現役／アーカイブ済みの絞り込みは 2 箇所（ページ付与の集計と
		// 本体）にあり、片方でも噛み合わないと、アーカイブ済みページに付与の事実が付かず、
		// 届いている本人にまで出なくなる（逆に絞りが緩めば、届いていない相手へ題名が出る）。
		// fake は SQL を通らないので、このずれは実 PostgreSQL でしか露見しない。
		f := setupKBPermission(t, sqlDB)
		secret := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "秘密")
		f.principalFor(ctx, t, f.alice) // alice も所属はしている（届く付与が無いだけ）
		bob := f.principalFor(ctx, t, f.bob)
		// bob にだけページ付与を張る（スペースの既定は誰にも張らない）。
		f.grantPage(ctx, t, secret.ID, bob.ID, domain.GrantRoleViewer)
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, secret.ID))

		assert.Empty(t, listFor(f, t, f.alice, true), "付与が届いていない相手には出ない")
		assert.Equal(t, []string{secret.ID}, listFor(f, t, f.bob, true), "届いている相手には出る")
	})

	t.Run("親がアーカイブ済みかを事実として返す", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "根")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "子")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, root.ID))

		out, err := usecase.NewListViewablePagesUseCase(f.perm).Execute(ctx,
			usecase.ListViewablePagesInput{
				WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.alice, Archived: true,
			})
		require.NoError(t, err)

		// 先に両方が一覧に出ていることを確かめる。map の引きは**鍵が無くても false** を返すので、
		// 根が欠落していても「復帰できる側」の検査だけは通ってしまう。
		assert.ElementsMatch(t, []string{root.ID, child.ID}, pageIDs(out.Pages))
		assert.False(t, out.ParentArchived[root.ID], "根は復帰できる側")
		assert.True(t, out.ParentArchived[child.ID], "子だけを復帰させることはできない")
	})
}

// ページを名指しする権限操作の入口が見る事実を確かめる。
//
// 入口（kbPermissionGate.requirePageAdmin）は「そのページに届いている既定の役割が admin か」で
// 判断する。役割はワークスペース / スペース / ページの 3 段のどこから来てもよく、
// 最も強いものが実効になる。ここで固定するのは、その 3 段が全部届くことと、
// 対象が引けない場合の返り方が実在で変わらないこと。
func TestKnowledgeBasePageManageFacts_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	canManage := func(t *testing.T, f kbPermFixture, pageID string, userID uint64) bool {
		t.Helper()
		facts, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, pageID, userID)
		require.NoError(t, err)
		return domain.ResolvePagePermission(*facts).CanManage
	}

	t.Run("スペースの admin はページも管理できる", func(t *testing.T) {
		// スペースを名指しで引いた答えと食い違わないこと。ここが割れると、
		// スペースの管理者がスペースの設定は変えられるのにページの共有は触れない、
		// といった説明できない状態になる。
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleAdmin)

		want, err := f.perm.SpacePermissionFactsForUser(ctx, f.ws, f.spaceA, f.alice)
		require.NoError(t, err)
		require.True(t, domain.ResolveScopePermission(*want).CanManage, "前提: スペースでは admin")

		assert.True(t, canManage(t, f, page.ID, f.alice), "ページ経由でも同じ答えになる")
	})

	t.Run("ワークスペースの役割も届く", func(t *testing.T) {
		// workspace_grants は配下の全スペース・全ページに届く。
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice := f.principalFor(ctx, t, f.alice)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)

		assert.True(t, canManage(t, f, page.ID, f.alice))
	})

	t.Run("スペースの全員宛ての役割も届く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		f.principalFor(ctx, t, f.alice) // 所属していないと「全員」は効かない
		everyone := f.everyoneOf(ctx, t, f.spaceA)
		f.grantSpace(ctx, t, f.spaceA, everyone.ID, domain.GrantRoleAdmin)

		assert.True(t, canManage(t, f, page.ID, f.alice))
	})

	t.Run("ページに張られた admin も届く", func(t *testing.T) {
		// 3 段目。ここが届かないと、ページに admin を与えられた本人が
		// その権限を一切行使できない（与えられるのに使えない）。
		f := setupKBPermission(t, sqlDB)
		parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親").ID
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent, "子").ID
		alice := f.principalFor(ctx, t, f.alice)
		f.grantPage(ctx, t, child, alice.ID, domain.GrantRoleAdmin)

		assert.True(t, canManage(t, f, child, f.alice), "張ったページは管理できる")
		assert.False(t, canManage(t, f, parent, f.alice), "親へは上がらない")
	})

	t.Run("editor では管理できない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleEditor)

		assert.False(t, canManage(t, f, page.ID, f.alice))
	})

	t.Run("引けないページはどれも同じセンチネルになる", func(t *testing.T) {
		// 応答の差から「そのページ ID が実在するか」を読ませない。存在しない ID も
		// UUID ですらない文字列も他テナントのページも、返るものが同じであること。
		f := setupKBPermission(t, sqlDB)
		f.principalFor(ctx, t, f.alice)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws,
			f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleAdmin)
		require.NoError(t, err)
		other := mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "他社のページ")

		for _, c := range []struct {
			name   string
			pageID string
		}{
			{"存在しない UUID", "0198a000-0000-7000-8000-0000000000ff"},
			{"UUID ですらない文字列", "not-a-uuid"},
			{"他テナントのページ", other.ID},
		} {
			_, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, c.pageID, f.alice)
			assert.ErrorIs(t, err, repository.ErrPageNotFound, c.name)
		}
	})

	t.Run("実在するが役割が無いページは拒否へ倒れる", func(t *testing.T) {
		// 上の 3 つと返り方（値かエラーか）は違うが、呼び出し側では同じ拒否に落ちる。
		// どちらも DB への問い合わせは 1 回なので、時間差からも区別できない。
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		f.principalFor(ctx, t, f.carol) // 所属はしているが役割を 1 つも持たない

		assert.False(t, canManage(t, f, page.ID, f.carol))
	})
}

func TestKnowledgeBasePermission_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("メンバー追加は冪等で所属判定に効く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)

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
		f := setupKBPermission(t, sqlDB)
		// alice を「別ワークスペース」のメンバーにして、その主体 ID を本命ワークスペースで使う。
		foreign, err := f.perm.EnsureUserPrincipal(ctx, f.otherWS, f.alice)
		require.NoError(t, err)

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, foreign.ID, domain.GrantRoleAdmin)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_space_grants_principal")

		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, foreign.ID, domain.GrantRoleAdmin)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_workspace_grants_principal")

		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		_, err = f.perm.UpsertPageGrant(ctx, f.ws, page.ID, foreign.ID, domain.GrantRoleAdmin)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_page_grants_principal")
	})

	t.Run("存在しないユーザーのprincipalは作れない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// 弾いているのは DB の FK。まず生の INSERT で制約が効いていることを確かめる。
		_, err := f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind, user_id)
			 VALUES (gen_random_uuid(), $1, 'user', 999999999)`, f.ws,
		)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principals_user")

		// repository はそれを ErrUserNotFound へ翻訳する。制約違反のまま上へ流すと
		// 「ユーザー ID を間違えた」という入力の誤りが HTTP の入口で 500 になり、
		// 呼び出し側が DB 障害と区別できない（再試行すべきだと誤解する）。
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, 999999999)
		require.ErrorIs(t, err, repository.ErrUserNotFound)
	})

	t.Run("グループ名の重複は一意制約として返る", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		_, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "重複する名前")
		require.NoError(t, err)

		// 名前はワークスペース内で一意（uq_principals_group_name）。同名が 2 つあると
		// 権限を張る先を人が選べない。ここも制約違反のままではなくセンチネルで返す。
		_, err = f.perm.CreateGroupPrincipal(ctx, f.ws, "重複する名前")
		require.ErrorIs(t, err, repository.ErrPrincipalGroupNameTaken)

		// 別ワークスペースなら同じ名前を使える（一意なのはワークスペース内だけ）。
		_, err = f.perm.CreateGroupPrincipal(ctx, f.otherWS, "重複する名前")
		require.NoError(t, err)
	})

	t.Run("ユーザーを消すと権限も消える", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
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
		f := setupKBPermission(t, sqlDB)
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
		f := setupKBPermission(t, sqlDB)

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
		f := setupKBPermission(t, sqlDB)
		_, err := f.db.Exec(
			`INSERT INTO principals (id, workspace_id, kind, space_id)
			 VALUES (gen_random_uuid(), $1, 'space_all', $2)`, f.ws, f.otherSpc,
		)
		requirePgError(t, err, sqlStateForeignKeyViolation, "fk_principals_space")
	})

	t.Run("既定はスペースのgrantで決まる", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanView, "grant が無ければ見えない")

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		got := f.permFor(ctx, t, page.ID, f.alice)
		assert.True(t, got.CanView)
		assert.False(t, got.CanEdit, "viewer は編集できない")

		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		assert.True(t, f.permFor(ctx, t, page.ID, f.alice).CanEdit, "同じ主体の grant は 1 行のまま更新される")

		grants, err := f.perm.ListSpaceGrants(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		require.Len(t, grants, 1, "upsert なので行は増えない")

		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, f.ws, f.spaceA, alice.ID))
		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanView, "剥がせば見えなくなる")
	})

	t.Run("ワークスペースのgrantは全スペースに効きスペースのgrantで降格しない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		pageA := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "A ルート")
		pageB := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceB, nil, "B ルート")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)

		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)
		assert.True(t, f.permFor(ctx, t, pageA.ID, f.alice).CanEdit, "スペース A に grant が無くても効く")
		assert.True(t, f.permFor(ctx, t, pageB.ID, f.alice).CanEdit, "スペース B にも効く")

		// スペース側で弱い役割を張っても、強い方（ワークスペースの admin）が採られる。
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.True(t, f.permFor(ctx, t, pageA.ID, f.alice).CanEdit,
			"スペースに viewer を張るだけでワークスペース管理者を締め出せてはいけない")

		// 取り消す前に別の admin を用意する。ユーザーの admin が 0 人になる取り消しは
		// repository が断るので、そこで落ちると本題（役割の合成規則）が確かめられない。
		keepAdmin(ctx, t, f, f.bob)

		require.NoError(t, f.perm.DeleteWorkspaceGrant(ctx, f.ws, alice.ID))
		got := f.permFor(ctx, t, pageA.ID, f.alice)
		assert.True(t, got.CanView, "スペースの viewer が残る")
		assert.False(t, got.CanEdit)
	})

	t.Run("グループ経由とスペース全員の権限が効く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		_, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		assert.True(t, f.permFor(ctx, t, page.ID, f.alice).CanView, "スペース全員の grant はメンバーに効く")
		assert.False(t, f.permFor(ctx, t, page.ID, f.carol).CanView, "非メンバーには効かない")

		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "開発")
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, group.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		bobPrincipal, err := f.perm.FindUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, bobPrincipal.ID))

		assert.True(t, f.permFor(ctx, t, page.ID, f.bob).CanEdit, "グループ経由で editor")
		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanEdit, "グループに入っていない人は viewer のまま")

		require.NoError(t, f.perm.RemoveGroupMember(ctx, f.ws, group.ID, bobPrincipal.ID))
		assert.False(t, f.permFor(ctx, t, page.ID, f.bob).CanEdit, "外せば既定に戻る")
	})

	t.Run("祖先のページ付与は子孫へ降り剥がせば入れ物の既定へ戻る", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")
		grand := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &child.ID, "grand")

		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleViewer)
		require.NoError(t, err)
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		assert.False(t, f.permFor(ctx, t, grand.ID, f.alice).CanEdit, "ページ付与が無ければ入れ物の既定どおり")

		// child に「alice を editor」。孫まで降りる。
		f.grantPage(ctx, t, child.ID, alice.ID, domain.GrantRoleEditor)
		assert.False(t, f.permFor(ctx, t, root.ID, f.alice).CanEdit, "祖先側（root）へは上がらない")
		assert.True(t, f.permFor(ctx, t, child.ID, f.alice).CanEdit)
		assert.True(t, f.permFor(ctx, t, grand.ID, f.alice).CanEdit, "子孫へ降りる")
		assert.False(t, f.permFor(ctx, t, grand.ID, f.bob).CanEdit, "張っていない相手の既定は動かない")
		assert.True(t, f.permFor(ctx, t, grand.ID, f.bob).CanView, "閲覧の既定（viewer）はそのまま")

		// 剥がすと入れ物の既定（スペース全員の viewer）へ戻る。
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, child.ID, alice.ID))
		assert.False(t, f.permFor(ctx, t, grand.ID, f.alice).CanEdit, "足した分だけが消える")
		assert.True(t, f.permFor(ctx, t, grand.ID, f.alice).CanView, "入れ物の既定は残る")
	})

	t.Run("近い段の付与はその枝だけを広げ遠い段の付与は木全体に効く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "child")
		grand := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &child.ID, "grand")

		// スペースの既定は誰にも張らない（届くのはページ付与だけ）。
		alice := f.principalFor(ctx, t, f.alice)
		bob := f.principalFor(ctx, t, f.bob)

		// root は alice に。child でその枝だけ bob にも広げる。
		f.grantPage(ctx, t, root.ID, alice.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, child.ID, bob.ID, domain.GrantRoleViewer)

		assert.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "root には bob へ届く付与が無い")
		assert.True(t, f.permFor(ctx, t, child.ID, f.bob).CanView, "child 以下だけ広げられる")
		assert.True(t, f.permFor(ctx, t, grand.ID, f.bob).CanView, "child の付与は子孫にも降りる")
		assert.ElementsMatch(t, []string{child.ID, grand.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))
		assert.ElementsMatch(t, []string{root.ID, child.ID, grand.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.alice))
	})

	t.Run("第三者へのページ付与を足しても届いていない人には見えないまま", func(t *testing.T) {
		// 付与は名指しした相手にだけ足される。誰かのために 1 行張ったことが、
		// 無関係な人へ波及しないことを固定する（波及すると、権限設定の 1 操作が
		// 意図しない相手にまで及ぶ）。
		f := setupKBPermission(t, sqlDB)
		parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "人事・機密")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent.ID, "査定シート")
		grand := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &child.ID, "評価コメント")

		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		carol := f.principalFor(ctx, t, f.carol)

		// この木は alice にだけ届く（スペースの既定は誰にも張らない）。
		f.grantPage(ctx, t, parent.ID, alice.ID, domain.GrantRoleEditor)
		for _, page := range []*domain.Page{parent, child, grand} {
			got := f.permFor(ctx, t, page.ID, f.bob)
			require.False(t, got.CanView, "付与が届いていない時点で bob には見えない")
			require.False(t, got.CanEdit)
		}
		require.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))

		// 「carol にだけ読ませる」という通常運用の付与を子ページへ 1 行足す。
		f.grantPage(ctx, t, child.ID, carol.ID, domain.GrantRoleViewer)

		for _, page := range []*domain.Page{child, grand} {
			got := f.permFor(ctx, t, page.ID, f.bob)
			assert.False(t, got.CanView, "第三者への付与 1 行で無関係な人に開いてはいけない")
			assert.False(t, got.CanEdit, "読めないページを編集できてもいけない")
		}
		assert.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.bob), "ツリー一覧にも露出しない")

		// 足した本人には届き、先に張ってあった相手の権限は下がらない。
		aliceOnChild := f.permFor(ctx, t, child.ID, f.alice)
		assert.True(t, aliceOnChild.CanView, "祖先の付与を持つ本人はそのまま")
		assert.True(t, aliceOnChild.CanEdit, "弱い付与を隣に足しても降格しない")
		assert.True(t, f.permFor(ctx, t, child.ID, f.carol).CanView, "名指しした本人には届く")
		assert.False(t, f.permFor(ctx, t, parent.ID, f.carol).CanView, "張った段より上へは上がらない")
		assert.ElementsMatch(t, []string{parent.ID, child.ID, grand.ID},
			f.viewablePageIDs(ctx, t, f.spaceA, f.alice), "祖先の付与は子孫まで効き続ける")
	})

	t.Run("主体を消してもほかの人の見え方は変わらない", func(t *testing.T) {
		// 引き金は攻撃ではなく通常運用（退職者のオフボーディング・部署の統廃合）。
		// 主体を消すと、その主体宛ての付与も FK の CASCADE で一緒に消える。消えるのは
		// **その人に届いていた分だけ**で、ほかの誰かに届く／届かないは 1 つも動かない。
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "人事・機密")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "査定シート")
		byGroup := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "部署だけの棚")

		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		f.principalFor(ctx, t, f.carol)
		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "人事部")
		require.NoError(t, err)

		f.grantPage(ctx, t, root.ID, alice.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, byGroup.ID, group.ID, domain.GrantRoleViewer)
		require.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "bob へ届く付与は最初から無い")
		require.False(t, f.permFor(ctx, t, byGroup.ID, f.bob).CanView)
		require.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))

		// 退職者を外す（付与が張られていた本人）。
		require.NoError(t, usecase.NewRemoveWorkspaceMemberUseCase(f.perm).Execute(ctx,
			usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}))
		// 部署の統廃合でグループを消す（付与が張られていた主体）。
		require.NoError(t, f.perm.DeletePrincipal(ctx, f.ws, group.ID))

		for _, page := range []*domain.Page{root, child, byGroup} {
			got := f.permFor(ctx, t, page.ID, f.bob)
			assert.False(t, got.CanView, "主体が消えても他人へ開かない: "+page.Title)
			assert.False(t, got.CanEdit, "読めないページを編集できてもいけない: "+page.Title)
		}
		assert.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.bob), "ツリー一覧にも出ない")
		assert.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.carol))

		// 主体ごと消えたので、その段には行が 1 つも残らない。
		rows, err := f.perm.ListPageGrants(ctx, f.ws, root.ID)
		require.NoError(t, err)
		assert.Empty(t, rows, "載っていた主体ごと付与の行は消えている")

		// 閉じたままにするのが目的で、開き直せなくなるわけではない。
		bob, err := f.perm.FindUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		f.grantPage(ctx, t, root.ID, bob.ID, domain.GrantRoleViewer)
		assert.True(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "張り直せば見える")
		assert.False(t, f.permFor(ctx, t, root.ID, f.carol).CanView, "張っていない人は見えないまま")
	})

	t.Run("段ごとに相手が違うとき上の段の主体を消しても下の段は動かない", func(t *testing.T) {
		// root = [alice] / child = [alice, bob] のように段ごとに相手が違うとき、
		// alice を消すと root の段だけが空になる。空になった段が「誰でも見える」に
		// 化けると、root 直下（child 以外）が全開になる。
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "広げた枝")
		sibling := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "広げていない枝")

		alice := f.principalFor(ctx, t, f.alice)
		bob := f.principalFor(ctx, t, f.bob)
		f.principalFor(ctx, t, f.carol)

		f.grantPage(ctx, t, root.ID, alice.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, child.ID, bob.ID, domain.GrantRoleViewer)
		require.ElementsMatch(t, []string{child.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))

		require.NoError(t, usecase.NewRemoveWorkspaceMemberUseCase(f.perm).Execute(ctx,
			usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}))

		assert.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "空になった段は全開にならない")
		assert.False(t, f.permFor(ctx, t, sibling.ID, f.bob).CanView, "root 直下の別の枝も閉じたまま")
		assert.True(t, f.permFor(ctx, t, child.ID, f.bob).CanView, "自分に張られた段はそのまま")
		assert.ElementsMatch(t, []string{child.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))
		assert.Empty(t, f.viewablePageIDs(ctx, t, f.spaceA, f.carol))
	})

	t.Run("見え方が変わるのはその人の付与を触ったときだけ", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")

		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		carol := f.principalFor(ctx, t, f.carol)

		f.grantPage(ctx, t, root.ID, alice.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, root.ID, carol.ID, domain.GrantRoleViewer)
		require.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView)

		// 他人の付与を剥がしても、自分の見え方は動かない。
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, root.ID, carol.ID))
		assert.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "他人の行を消しても開かない")
		assert.True(t, f.permFor(ctx, t, root.ID, f.alice).CanView, "残っている本人はそのまま")
		assert.False(t, f.permFor(ctx, t, root.ID, f.carol).CanView, "剥がした本人は見えなくなる")

		// 最後の 1 行を消しても、届く段がほかに無いので誰にも開かない。
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, root.ID, alice.ID))
		assert.False(t, f.permFor(ctx, t, root.ID, f.alice).CanView)
		assert.False(t, f.permFor(ctx, t, root.ID, f.bob).CanView, "行が 0 になっても全開にはならない")
		rows, err := f.perm.ListPageGrants(ctx, f.ws, root.ID)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("編集できるのは編集の付与が届いた人だけ", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "規程集")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "就業規則")

		// スペース全員には viewer（読み取り専用のサブツリー）。
		f.grantSpace(ctx, t, f.spaceA, f.everyoneOf(ctx, t, f.spaceA).ID, domain.GrantRoleViewer)
		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		carol := f.principalFor(ctx, t, f.carol)

		// root 以下は「alice だけが編集できる」。
		f.grantPage(ctx, t, root.ID, alice.ID, domain.GrantRoleEditor)
		require.True(t, f.permFor(ctx, t, child.ID, f.bob).CanView, "閲覧の既定は viewer のまま")
		require.False(t, f.permFor(ctx, t, child.ID, f.bob).CanEdit)

		// 「carol にも読ませる」つもりの付与を子に 1 行足す。
		f.grantPage(ctx, t, child.ID, carol.ID, domain.GrantRoleViewer)

		assert.False(t, f.permFor(ctx, t, child.ID, f.bob).CanEdit,
			"読み取り専用サブツリーが全員に開いてはいけない（データ破壊になる）")
		assert.True(t, f.permFor(ctx, t, child.ID, f.alice).CanEdit, "編集を張られた本人はそのまま")
		assert.False(t, f.permFor(ctx, t, child.ID, f.carol).CanEdit, "viewer を足しても編集にはならない")
		assert.True(t, f.permFor(ctx, t, child.ID, f.carol).CanView)
	})

	t.Run("一覧は別ワークスペースのスペースとアーカイブ済みを返さない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		leaving := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "アーカイブする子")
		f.grantSpace(ctx, t, f.spaceA, f.principalFor(ctx, t, f.alice).ID, domain.GrantRoleEditor)
		require.ElementsMatch(t, []string{root.ID, leaving.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.alice))

		require.NoError(t, f.pageUC.archive.Execute(ctx, usecase.ArchivePageInput{
			WorkspaceID: f.ws, PageID: leaving.ID,
		}))
		assert.ElementsMatch(t, []string{root.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.alice),
			"アーカイブ済みは一覧に出ない")

		// 別ワークスペースのスペース ID を渡しても 1 枚も返さない（事実の収集の時点で塞ぐ）。
		mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "別テナントのページ")
		foreign, err := f.perm.ListSpacePageViewFacts(ctx, f.ws, f.otherSpc, f.alice, false)
		require.NoError(t, err)
		assert.Empty(t, foreign, "テナント越えの spaceID では 0 件")
	})

	t.Run("スペース全員宛てのページ付与は別スペースへの移動で失効しない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親")
		shared := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent.ID, "全員に配った子")

		f.principalFor(ctx, t, f.bob)
		everyoneA := f.everyoneOf(ctx, t, f.spaceA)
		f.grantPage(ctx, t, shared.ID, everyoneA.ID, domain.GrantRoleEditor)
		require.True(t, f.permFor(ctx, t, shared.ID, f.bob).CanEdit, "前提: スペース A の全員に届いている")

		// 付与を持つページの祖先を、別スペースのルートへ動かす（正規の操作）。
		// 移動後は「スペース A の全員」が対象外になり、行だけが残って効かなくなる。
		_, err := f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.Error(t, err, "見えている行が黙って効かなくなる移動は失敗させる")
		assert.True(t, f.permFor(ctx, t, shared.ID, f.bob).CanEdit, "移動していないので付与は効いたまま")

		require.ErrorIs(t, err, repository.ErrPageMoveVoidsSpaceGrant)
		moved, err := f.pages.FindPage(ctx, f.ws, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, f.spaceA, moved.SpaceID, "失敗した移動はロールバックされている")

		// 役割が違っても同じ扱い（強い付与だけ止める、という非対称を残さない）。
		f.grantPage(ctx, t, shared.ID, everyoneA.ID, domain.GrantRoleViewer)
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.ErrorIs(t, err, repository.ErrPageMoveVoidsSpaceGrant, "viewer でも同じ扱い")

		// 付与を先に整理すれば移せる（止めるのは「意味を失う付与が残っているとき」だけ）。
		require.NoError(t, f.perm.DeletePageGrant(ctx, f.ws, shared.ID, everyoneA.ID))
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: parent.ID, NewSpaceID: f.spaceB,
		})
		require.NoError(t, err)
		moved, err = f.pages.FindPage(ctx, f.ws, parent.ID)
		require.NoError(t, err)
		assert.Equal(t, f.spaceB, moved.SpaceID)
	})

	t.Run("移動先スペース全員の付与と同一スペース内の移動は止めない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)

		// スペースが変わらない移動は、付与の意味も変わらないので止めない。
		staying := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親")
		newParent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "別の親")
		stayingChild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &staying.ID, "全員に配った子")
		f.grantPage(ctx, t, stayingChild.ID, f.everyoneOf(ctx, t, f.spaceA).ID, domain.GrantRoleViewer)
		_, err := f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: staying.ID, NewParentID: &newParent.ID,
		})
		require.NoError(t, err, "同一スペース内の移動は付与の意味を変えない")

		// 「移動先スペースの全員」宛ての付与は、移動後にこそ意味を持つので止めない。
		leaving := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "移すページ")
		leavingChild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &leaving.ID, "移す子")
		f.grantPage(ctx, t, leavingChild.ID, f.everyoneOf(ctx, t, f.spaceB).ID, domain.GrantRoleViewer)
		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: leaving.ID, NewSpaceID: f.spaceB,
		})
		require.NoError(t, err, "移動先スペース宛ての付与は移動後に効くので止めない")
	})

	t.Run("祖先をたどる解決が別スペースへ漏れない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// スペース A と B に、それぞれ独立した木を作る。
		aRoot := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "A ルート")
		aChild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &aRoot.ID, "A 子")
		bRoot := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceB, nil, "B ルート")

		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		// スペース B だけ全員 editor。A には入れ物の既定を置かない。
		everyoneB, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceB)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceB, everyoneB.ID, domain.GrantRoleEditor)
		require.NoError(t, err)

		// A ルートに alice への付与を張っても、B の木には 1 つも影響しない。
		f.grantPage(ctx, t, aRoot.ID, alice.ID, domain.GrantRoleViewer)
		assert.True(t, f.permFor(ctx, t, aChild.ID, f.alice).CanView, "A の木は張った本人にだけ降りる")
		assert.False(t, f.permFor(ctx, t, aChild.ID, f.bob).CanView, "A の木は bob へ届かない")
		assert.True(t, f.permFor(ctx, t, bRoot.ID, f.bob).CanView, "B の木は無関係のまま")
		assert.True(t, f.permFor(ctx, t, bRoot.ID, f.alice).CanView,
			"B は全員 editor なので alice にも見える")

		// スペース全員の grant もスペースごとに独立している。
		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, f.ws, f.spaceB, everyoneB.ID))
		assert.False(t, f.permFor(ctx, t, bRoot.ID, f.bob).CanView)
		assert.True(t, f.permFor(ctx, t, aChild.ID, f.alice).CanView, "B の grant を剥がしても A は変わらない")
	})

	t.Run("ページを別のスペースへ動かすと実効権限が変わる", func(t *testing.T) {
		// 見せたくないものは private のスペースへ置く、という運用をそのまま通す。
		//
		// ここで見るのは**スペースをまたぐ移動**だけ（下の move はどちらも NewSpaceID を渡す）。
		// 同じスペースの中で親を替えても見え方が変わらないことは、
		// TestPageGrant_木を下るほど役割は弱くならない_Integration が単調性として押さえている。
		f := setupKBPermission(t, sqlDB)
		open := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "公開の親")
		moving := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &open.ID, "動くページ")

		// スペース A は全員 editor。スペース B は private で alice だけ。
		everyone, err := f.perm.EnsureSpaceEveryonePrincipal(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, everyone.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		f.makePrivate(t, f.spaceB)
		f.grantSpace(ctx, t, f.spaceB, alice.ID, domain.GrantRoleEditor)

		assert.True(t, f.permFor(ctx, t, moving.ID, f.bob).CanView, "スペース A では bob も見える")

		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: moving.ID, NewSpaceID: f.spaceB,
		})
		require.NoError(t, err)
		assert.False(t, f.permFor(ctx, t, moving.ID, f.bob).CanView, "private のスペースへ移すと見えなくなる")
		assert.True(t, f.permFor(ctx, t, moving.ID, f.alice).CanView, "そのスペースへ張られた本人には見える")

		_, err = f.pageUC.move.Execute(ctx, usecase.MovePageInput{
			WorkspaceID: f.ws, PageID: moving.ID, NewSpaceID: f.spaceA,
		})
		require.NoError(t, err)
		assert.True(t, f.permFor(ctx, t, moving.ID, f.bob).CanView, "戻せばまた見える（付与の行は 1 つも触っていない）")
	})

	t.Run("閲覧可能ページ一覧は届いていないページを落とす", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		open := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "公開")
		secret := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "秘密")
		secretChild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &secret.ID, "秘密の子")

		// 入れ物の既定は誰にも張らない。届くのはページ付与だけ。
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		bob, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
		require.NoError(t, err)
		f.grantPage(ctx, t, open.ID, bob.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, open.ID, alice.ID, domain.GrantRoleViewer)
		f.grantPage(ctx, t, secret.ID, alice.ID, domain.GrantRoleViewer)

		listUC := usecase.NewListViewablePagesUseCase(f.perm)
		bobPages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.bob,
		})
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{open.ID}, pageIDs(bobPages.Pages), "秘密の木は丸ごと落ちる")

		alicePages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.alice,
		})
		require.NoError(t, err)
		assert.ElementsMatch(t,
			[]string{open.ID, secret.ID, secretChild.ID}, pageIDs(alicePages.Pages),
			"付与された人には子孫まで見える")

		carolPages, err := listUC.Execute(ctx, usecase.ListViewablePagesInput{
			WorkspaceID: f.ws, SpaceID: f.spaceA, UserID: f.carol,
		})
		require.NoError(t, err)
		assert.Empty(t, carolPages.Pages, "非メンバーには 1 枚も見えない")
		assert.Empty(t, carolPages.HasHiddenChildren, "印も返さない（実在が漏れる）")
	})

	t.Run("一覧は所属グループとスペース全員の付与を1ページ解決と同じに畳む", func(t *testing.T) {
		// 一覧は 1 ページの解決とは別に書かれた同型の集計で、片方だけ壊れても
		// もう片方のテストでは気づけない。自分宛ての付与しか置かない配役では、
		// 一覧側の所属グループ・スペース全員の枝を落としても素通りするため、
		// 3 つの主体すべてが一覧経路にも効いていることをここで固定する。
		f := setupKBPermission(t, sqlDB)
		mine := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "自分宛てのページ")
		byGroup := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "部署宛てのページ")
		byGroupChild := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &byGroup.ID, "その子")
		byEveryone := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "全員宛てのページ")
		unreachable := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "誰にも張っていないページ")

		alice := f.principalFor(ctx, t, f.alice)
		f.principalFor(ctx, t, f.bob)
		group, err := f.perm.CreateGroupPrincipal(ctx, f.ws, "総務")
		require.NoError(t, err)
		require.NoError(t, f.perm.AddGroupMember(ctx, f.ws, group.ID, alice.ID))

		// 自分宛て。
		f.grantPage(ctx, t, mine.ID, alice.ID, domain.GrantRoleViewer)
		// 所属グループ宛て（自分の主体だけを見ていると一覧で無視される）。
		f.grantPage(ctx, t, byGroup.ID, group.ID, domain.GrantRoleViewer)
		// スペース全員宛て（所属している人にだけ届く）。
		f.grantPage(ctx, t, byEveryone.ID, f.everyoneOf(ctx, t, f.spaceA).ID, domain.GrantRoleViewer)

		aliceViewable := f.viewablePageIDs(ctx, t, f.spaceA, f.alice)
		assert.ElementsMatch(t,
			[]string{mine.ID, byGroup.ID, byGroupChild.ID, byEveryone.ID}, aliceViewable,
			"自分・所属グループ・スペース全員の 3 経路とも一覧に効く")

		// 1 ページずつの解決と一覧が同じ答えになること（別々に書かれた集計なので突き合わせる）。
		for _, page := range []*domain.Page{mine, byGroup, byGroupChild, byEveryone, unreachable} {
			assert.Equal(t, f.permFor(ctx, t, page.ID, f.alice).CanView,
				slices.Contains(aliceViewable, page.ID), "1 ページ解決と一覧が割れている: "+page.Title)
		}

		// bob はグループに入っていないので、全員宛ての 1 枚だけが見える。
		assert.ElementsMatch(t, []string{byEveryone.ID}, f.viewablePageIDs(ctx, t, f.spaceA, f.bob))

		// グループから外すと、グループ宛ての付与は届かなくなる。
		require.NoError(t, f.perm.RemoveGroupMember(ctx, f.ws, group.ID, alice.ID))
		assert.ElementsMatch(t, []string{mine.ID, byEveryone.ID},
			f.viewablePageIDs(ctx, t, f.spaceA, f.alice), "所属が消えればグループ宛ての付与も届かない")
	})

	t.Run("メンバー削除のusecaseは主体ごと消す", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		_, err = f.perm.UpsertSpaceGrant(ctx, f.ws, f.spaceA, alice.ID, domain.GrantRoleEditor)
		require.NoError(t, err)
		require.True(t, f.permFor(ctx, t, page.ID, f.alice).CanEdit)

		removeUC := usecase.NewRemoveWorkspaceMemberUseCase(f.perm)
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}))
		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanView, "所属を外すと権限も消える")
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.alice}),
			"二度目は冪等に成功する")

		require.ErrorIs(t, f.perm.DeletePrincipal(ctx, f.ws, alice.ID), repository.ErrPrincipalNotFound)
	})

	t.Run("3段のgrantをusecase経由で張って一覧を引ける", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		alice, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.alice)
		require.NoError(t, err)
		bob, err := f.perm.EnsureUserPrincipal(ctx, f.ws, f.bob)
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

		// 3 段目（ページ）。bob はここで初めて役割を得る。
		grantPageUC := usecase.NewGrantPageRoleUseCase(f.perm)
		_, err = grantPageUC.Execute(ctx, usecase.GrantPageRoleInput{
			WorkspaceID: f.ws, PageID: page.ID, PrincipalID: bob.ID, Role: domain.GrantRoleEditor,
		})
		require.NoError(t, err)
		pageGrants, err := usecase.NewListPageGrantsUseCase(f.perm).Execute(ctx,
			usecase.ListPageGrantsInput{WorkspaceID: f.ws, PageID: page.ID})
		require.NoError(t, err)
		require.Len(t, pageGrants, 1)
		assert.Equal(t, domain.GrantRoleEditor, pageGrants[0].Role)
		assert.True(t, f.permFor(ctx, t, page.ID, f.bob).CanEdit, "ページ付与で編集できる")
		assert.False(t, f.permFor(ctx, t, page.ID, f.bob).CanManage, "editor では権限を変えられない")

		require.NoError(t, usecase.NewRevokePageRoleUseCase(f.perm).Execute(ctx,
			usecase.RevokePageRoleInput{WorkspaceID: f.ws, PageID: page.ID, PrincipalID: bob.ID}))
		assert.False(t, f.permFor(ctx, t, page.ID, f.bob).CanView, "剥がせば届かない")

		// 取り消す前に別の admin を用意する（0 人になる取り消しは repository が断る）。
		keepAdmin(ctx, t, f, f.carol)
		require.NoError(t, usecase.NewRevokeWorkspaceRoleUseCase(f.perm).Execute(ctx,
			usecase.RevokeWorkspaceRoleInput{WorkspaceID: f.ws, PrincipalID: alice.ID}))
		assert.False(t, f.permFor(ctx, t, page.ID, f.alice).CanView)
	})

	t.Run("グループ操作のusecaseが権限に効く", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		bobPrincipal, err := usecase.NewAddWorkspaceMemberUseCase(f.perm).Execute(ctx,
			usecase.AddWorkspaceMemberInput{WorkspaceID: f.ws, UserID: f.bob})
		require.NoError(t, err)
		// メンバー追加は既定で editor を付ける。この試験は「グループ経由の権限」だけを
		// 見たいので、既定の役割を外して素の状態（役割なしのメンバー）から始める。
		require.NoError(t, usecase.NewRevokeWorkspaceRoleUseCase(f.perm).Execute(ctx,
			usecase.RevokeWorkspaceRoleInput{WorkspaceID: f.ws, PrincipalID: bobPrincipal.ID}))
		assert.False(t, f.permFor(ctx, t, page.ID, f.bob).CanView, "役割を外した直後は見えない")
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
		assert.True(t, f.permFor(ctx, t, page.ID, f.bob).CanEdit)

		removeUC := usecase.NewRemoveGroupMemberUseCase(f.perm)
		require.NoError(t, removeUC.Execute(ctx, usecase.RemoveGroupMemberInput{
			WorkspaceID: f.ws, GroupPrincipalID: group.ID, MemberUserID: f.bob,
		}))
		assert.False(t, f.permFor(ctx, t, page.ID, f.bob).CanView)

		// スペース全員の主体も usecase 経由で用意でき、二度呼んでも増えない。
		everyoneUC := usecase.NewEnsureSpaceEveryonePrincipalUseCase(f.perm)
		first, err := everyoneUC.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{WorkspaceID: f.ws, SpaceID: f.spaceA})
		require.NoError(t, err)
		second, err := everyoneUC.Execute(ctx, usecase.EnsureSpaceEveryonePrincipalInput{WorkspaceID: f.ws, SpaceID: f.spaceA})
		require.NoError(t, err)
		assert.Equal(t, first.ID, second.ID)
	})

	t.Run("形式が不正なIDは存在しないものとして扱う", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
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
		pageGrants, err := f.perm.ListPageGrants(ctx, bad, bad)
		require.NoError(t, err)
		assert.Empty(t, pageGrants)
		links, err := f.perm.ListPageShareLinks(ctx, bad, bad)
		require.NoError(t, err)
		assert.Empty(t, links)
		facts, err := f.perm.ListSpacePageViewFacts(ctx, bad, bad, f.alice, false)
		require.NoError(t, err)
		assert.Empty(t, facts)

		require.NoError(t, f.perm.RemoveGroupMember(ctx, bad, bad, bad))
		require.NoError(t, f.perm.DeleteWorkspaceGrant(ctx, bad, bad))
		require.NoError(t, f.perm.DeleteSpaceGrant(ctx, bad, bad, bad))
		require.NoError(t, f.perm.DeletePageGrant(ctx, bad, bad, bad))
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
		_, err = f.perm.UpsertPageGrant(ctx, bad, bad, bad, domain.GrantRoleViewer)
		require.ErrorIs(t, err, repository.ErrPrincipalNotFound)
		_, err = f.perm.CreateShareLink(ctx, repository.ShareLinkWrite{WorkspaceID: bad, PageID: bad})
		require.ErrorIs(t, err, repository.ErrPageNotFound)
	})

	t.Run("別テナントの所属は解決に持ち込まれない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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
		f := setupKBPermission(t, sqlDB)
		foreignPage := mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "別テナントのページ")

		_, err := f.perm.PagePermissionFactsForUser(ctx, f.ws, foreignPage.ID, f.alice)
		require.ErrorIs(t, err, repository.ErrPageNotFound, "テナント越えは「無い」と同じ扱い")
	})
}

func TestKnowledgeBaseShareLink_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("発行したリンクで対象ページと子孫を開ける", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "子")
		outside := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "対象外")

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

	t.Run("リンクの主体に付与を張っても権限は変えられない", func(t *testing.T) {
		// 共有リンクの来訪者はログインしていない。その相手が「権限を変える」側に回れると、
		// URL を知っているだけの人が誰に何を見せるかを決められることになる。
		//
		// **この状態は API から作れる。** 付与の口は主体の実在しか確かめず、種類を見ない。
		// 共有リンクの主体 ID はリンク一覧の応答に載るので、admin が取り違えて張れてしまう。
		// 規則の側（ResolvePagePermission）で閉じていることを、実 DB の経路で固定する。
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "公開ページ")

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: page.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)

		// リンクの主体へ admin を張る（本来やるべきではない操作だが、いまは通ってしまう）。
		f.grantPage(ctx, t, page.ID, issued.Link.PrincipalID, domain.GrantRoleAdmin)

		verified, err := usecase.NewVerifyShareLinkUseCase(f.perm).
			Execute(ctx, usecase.VerifyShareLinkInput{Token: issued.Token})
		require.NoError(t, err)

		got, err := usecase.NewCheckShareLinkPermissionUseCase(f.perm, f.pages).
			Execute(ctx, usecase.CheckShareLinkPermissionInput{Link: verified, PageID: page.ID})
		require.NoError(t, err)

		assert.False(t, got.CanManage,
			"リンクの来訪者が権限を変えられてはいけない（付与を張られても）")
		assert.False(t, got.CanEdit, "閲覧のリンクなので編集もできない")
		assert.True(t, got.CanView, "閲覧はリンクの既定どおりできる")
	})

	t.Run("平文トークンはDBに残らない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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

	t.Run("リンクの来訪者の見え方はメンバーへの付与で変わらない", func(t *testing.T) {
		// 共有リンクは広げる方向にしか働かない。来訪者にできることはリンクの capability
		// だけで決まり、メンバーへ張った付与（3 段のどれでも）はそこへ足し引きしない。
		// 逆向きも同じで、来訪者を弱める手段は無い（弱める層をどこにも持たない）。
		f := setupKBPermission(t, sqlDB)
		root := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "公開ルート")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &root.ID, "子")

		issued, err := usecase.NewIssueShareLinkUseCase(f.perm).Execute(ctx, usecase.IssueShareLinkInput{
			WorkspaceID: f.ws, PageID: root.ID, Capability: domain.CapabilityView, CreatedByUserID: f.alice,
		})
		require.NoError(t, err)
		linkPerm := shareLinkPermFunc(ctx, t, f)

		before := linkPerm(issued.Token, child.ID)
		require.True(t, before.CanView, "前提: 閲覧のリンクで子まで開ける")
		require.False(t, before.CanEdit)

		// メンバーの権限を厚くする（ワークスペース admin / スペース全員 admin /
		// ページ admin の 3 段すべて）。来訪者はこれを 1 つも拾わない。
		bob := f.principalFor(ctx, t, f.bob)
		_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, bob.ID, domain.GrantRoleAdmin)
		require.NoError(t, err)
		f.grantSpace(ctx, t, f.spaceA, f.everyoneOf(ctx, t, f.spaceA).ID, domain.GrantRoleAdmin)
		f.grantPage(ctx, t, child.ID, bob.ID, domain.GrantRoleAdmin)

		after := linkPerm(issued.Token, child.ID)
		assert.True(t, after.CanView, "閲覧はリンクの capability のまま")
		assert.False(t, after.CanEdit, "未認証の来訪者にメンバーの admin が渡ってはいけない")
		assert.False(t, after.CanManage, "権限を変えられる側へは絶対に倒さない")
	})

	t.Run("共有リンクの主体はスペースのgrantを拾わない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
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
		f := setupKBPermission(t, sqlDB)
		page := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "root")
		other := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "別ページ")
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
func shareLinkPermFunc(ctx context.Context, t *testing.T, f kbPermFixture) func(token, pageID string) domain.PagePermission {
	t.Helper()
	verifyUC := usecase.NewVerifyShareLinkUseCase(f.perm)
	checkUC := usecase.NewCheckShareLinkPermissionUseCase(f.perm, f.pages)
	return func(token, pageID string) domain.PagePermission {
		t.Helper()
		link, err := verifyUC.Execute(ctx, usecase.VerifyShareLinkInput{Token: token})
		require.NoError(t, err)
		got, err := checkUC.Execute(ctx,
			usecase.CheckShareLinkPermissionInput{Link: link, PageID: pageID})
		require.NoError(t, err)
		return *got
	}
}

// keepAdmin は userID をワークスペースの admin にする。
//
// 「最後の admin は外せない」は repository が書き込みと同じトランザクションで守っている。
// admin の取り消しそのものが本題でないテストは、先に 2 人目を用意してからでないと
// その検査に引っかかって、確かめたかったこと（役割の合成規則など）へ辿り着けない。
func keepAdmin(ctx context.Context, t *testing.T, f kbPermFixture, userID uint64) {
	t.Helper()
	p, err := f.perm.EnsureUserPrincipal(ctx, f.ws, userID)
	require.NoError(t, err)
	_, err = f.perm.UpsertWorkspaceGrant(ctx, f.ws, p.ID, domain.GrantRoleAdmin)
	require.NoError(t, err)
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

func TestKnowledgeBaseSearchViewFacts_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	searchFor := func(f kbPermFixture, t *testing.T, userID uint64, query string) []string {
		t.Helper()
		pages, err := usecase.NewSearchViewablePagesUseCase(f.perm).Execute(ctx,
			usecase.SearchViewablePagesInput{WorkspaceID: f.ws, UserID: userID, Query: query})
		require.NoError(t, err)
		return pageIDs(pages)
	}

	t.Run("題名の部分一致でスペースを跨いで返り、届いていないページは出ない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// スペース C は private。ワークスペース全体の付与もスペース全員の付与も届かず、
		// そのスペースを名指しした付与だけが届く（＝ 見せないための置き場）。
		spaceC := createSpace(t, sqlDB, f.ws, "ccc-private")
		f.makePrivate(t, spaceC)

		inA := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "Docker 手順")
		inB := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceB, nil, "docker 入門")
		secret := mustCreatePage(ctx, t, f.pageUC, f.ws, spaceC, nil, "Docker 機密")
		_ = mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "無関係")

		alice := f.principalFor(ctx, t, f.alice)
		bob := f.principalFor(ctx, t, f.bob)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
		f.grantSpace(ctx, t, f.spaceB, alice.ID, domain.GrantRoleViewer)
		f.grantSpace(ctx, t, f.spaceA, bob.ID, domain.GrantRoleViewer)
		// 機密のスペースへ入れるのは bob だけ（1 ページ解決と同じ見方で、検索にも効くこと）。
		f.grantSpace(ctx, t, spaceC, bob.ID, domain.GrantRoleViewer)

		got := searchFor(f, t, f.alice, "docker")
		// 並びは題名順（"Docker 手順" < "docker 入門" は ILIKE ではなく ORDER BY title 依存）。
		// 順序はロケールに寄るので、集合として確かめる。
		assert.ElementsMatch(t, []string{inA.ID, inB.ID}, got)
		// bob は private スペースへ入れるので機密も出る。一方 B の付与は無いので
		// B のページは出ない（スペースごとの権限が検索でも効いている確認を兼ねる）。
		assert.ElementsMatch(t, []string{inA.ID, secret.ID}, searchFor(f, t, f.bob, "docker"))
	})

	t.Run("付与の届かない木は検索に出ない", func(t *testing.T) {
		// 一覧（木）では届いていない枝の中身は出ない。検索が別の判定を持つと
		// 「木には出ないのに検索では出る」穴になる — 同じ ResolvePageView を通る確認。
		f := setupKBPermission(t, sqlDB)
		parent := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "親")
		child := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, &parent.ID, "Docker 子")
		alice := f.principalFor(ctx, t, f.alice)
		// alice が持つのは別スペースの付与だけ（この木へは 1 つも届かない）。
		f.grantSpace(ctx, t, f.spaceB, alice.ID, domain.GrantRoleViewer)

		assert.Empty(t, searchFor(f, t, f.alice, "docker"), "届いていない木の子 %s が検索に出ている", child.ID)

		// 子にだけページ付与を張ると、その 1 枚だけが出る（親は出ない）。
		f.grantPage(ctx, t, child.ID, alice.ID, domain.GrantRoleViewer)
		assert.Equal(t, []string{child.ID}, searchFor(f, t, f.alice, "docker"),
			"ページ付与は検索の経路にも効く")
	})

	t.Run("ワークスペースの境界を越えない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		mine := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "共通の題名")
		_ = mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "共通の題名")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)

		assert.Equal(t, []string{mine.ID}, searchFor(f, t, f.alice, "共通"))
	})

	t.Run("LIKE の記号は文字として扱う（% で全件は返らない）", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		literal := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "進捗 100% の報告")
		_ = mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "無関係")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)

		assert.Equal(t, []string{literal.ID}, searchFor(f, t, f.alice, "100%"),
			"% がワイルドカードのまま渡ると全件一致になる")
	})

	t.Run("アーカイブ済みは検索に出ない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		gone := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "Docker 旧版")
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, gone.ID))

		assert.Empty(t, searchFor(f, t, f.alice, "docker"))
	})
}

func TestKnowledgeBaseUpdateSpaceName_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("名前だけが変わり key は変わらない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		require.NoError(t, f.pages.UpdateSpaceName(ctx, f.ws, f.spaceA, "改組後の名前"))
		sp, err := f.pages.FindSpace(ctx, f.ws, f.spaceA)
		require.NoError(t, err)
		assert.Equal(t, "改組後の名前", sp.Name)
		assert.Equal(t, "aaa", sp.Key)
	})

	t.Run("別ワークスペースのスペース ID は not found", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		err := f.pages.UpdateSpaceName(ctx, f.ws, f.otherSpc, "越境")
		assert.ErrorIs(t, err, repository.ErrSpaceNotFound)
		// 相手側の名前が変わっていないこと（0 件更新の確認を裏からも取る）。
		sp, ferr := f.pages.FindSpace(ctx, f.otherWS, f.otherSpc)
		require.NoError(t, ferr)
		assert.NotEqual(t, "越境", sp.Name)
	})
}

// ページ参照の題名解決に使う ID 指定の可視事実。検索と同じ規則で判定されることと、
// 境界（他ワークスペース・不正 ID・private のスペース）を実 PostgreSQL で確かめる。
func TestKnowledgeBaseViewFactsByIDs_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	ctx := context.Background()

	t.Run("指定IDの現役ページだけが返り、届いていないページは閲覧不可で載る", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// 機密は private のスペースへ置く。alice はワークスペース全体の付与を持つが、
		// private のスペースにはそれが届かない（この口が visibility を見ている確認）。
		secretSpace := createSpace(t, sqlDB, f.ws, "secret")
		f.makePrivate(t, secretSpace)

		visible := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "見えるページ")
		secret := mustCreatePage(ctx, t, f.pageUC, f.ws, secretSpace, nil, "機密ページ")
		other := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "頼んでいないページ")
		archived := mustCreatePage(ctx, t, f.pageUC, f.ws, f.spaceA, nil, "アーカイブしたページ")
		require.NoError(t, f.pages.ArchivePageSubtree(ctx, f.ws, archived.ID))

		alice := f.principalFor(ctx, t, f.alice)
		_, err := f.perm.UpsertWorkspaceGrant(ctx, f.ws, alice.ID, domain.GrantRoleViewer)
		require.NoError(t, err)

		rows, err := f.perm.ListWorkspacePageViewFactsByIDs(ctx, f.ws, f.alice,
			[]string{visible.ID, secret.ID, archived.ID, "not-a-uuid"})
		require.NoError(t, err)
		require.Len(t, rows, 3,
			"頼んだ ID のページが返る（不正 ID は静かに落ちる。アーカイブ済みは行として返り、除外は用途側の判断）")

		byID := map[string]bool{}
		archivedAt := map[string]bool{}
		for _, row := range rows {
			byID[row.Page.ID] = domain.ResolvePageView(row.Role)
			archivedAt[row.Page.ID] = row.Page.ArchivedAt != nil
		}
		assert.True(t, byID[visible.ID], "付与が届くページは閲覧できる")
		assert.False(t, byID[secret.ID],
			"private のスペースへワークスペース全体の付与は届かない（検索と同じ規則）")
		assert.True(t, archivedAt[archived.ID], "アーカイブ済みは ArchivedAt 付きで返る（呼び出し側が除外を判断できる）")
		assert.False(t, archivedAt[visible.ID])
		// 頼んでいない ID は返らない（ID で絞る口が全件の口にならない証拠）。
		_, unrequested := byID[other.ID]
		assert.False(t, unrequested)
	})

	t.Run("他ワークスペースのIDは返らない", func(t *testing.T) {
		f := setupKBPermission(t, sqlDB)
		// alice は f.ws の正規メンバー（境界を跨げないことを、権限がある状態で確かめる）。
		// fixture をもう 1 つ作らないのは、setupKBPermission が先頭で TruncateAll を
		// 実行し、先に作った fixture の行が全部消えてしまうため。
		alice := f.principalFor(ctx, t, f.alice)
		f.grantSpace(ctx, t, f.spaceA, alice.ID, domain.GrantRoleViewer)
		foreign := mustCreatePage(ctx, t, f.pageUC, f.otherWS, f.otherSpc, nil, "よそのページ")

		rows, err := f.perm.ListWorkspacePageViewFactsByIDs(ctx, f.ws, f.alice, []string{foreign.ID})
		require.NoError(t, err)
		assert.Empty(t, rows, "ワークスペース境界を跨いで題名を引けない")
	})
}
