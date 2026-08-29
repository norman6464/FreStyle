//go:build integration

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKnowledgeBaseWorkspaceAPI_Integration はワークスペース / スペースの入口を
// 実 PostgreSQL・本番と同じ配線で確かめる。
//
// このチケットの主眼は「API が一周すること」なので、途中を fake で埋めずに
// 作成 → 発見 → スペース作成 → 空のスペースへの最初のページ作成 まで通しで回す。
func TestKnowledgeBaseWorkspaceAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("作成から空のスペースへの最初のページ作成までAPIだけで一周する", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		e := env.as(alice)

		// 1. ワークスペースを作る。作成者は同じトランザクションで admin のメンバーになる。
		created := e.do(t, http.MethodPost, "/api/v2/kb/workspaces",
			`{"slug":"startup","name":"新会社"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())

		// 2. slug を一覧で発見できる（URL に slug を使う以上、知る手段が要る）。
		listed := e.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, listed.Code)
		var workspaces []kbWorkspaceResponse
		require.NoError(t, json.Unmarshal(listed.Body.Bytes(), &workspaces))
		require.Len(t, workspaces, 1, "自分が作ったものだけが見える")
		assert.Equal(t, "startup", workspaces[0].Slug)

		// 3. スペースを作る（作成者は admin なので通る）。
		spacesPath := "/api/v2/kb/workspaces/startup/spaces"
		spaceRes := e.do(t, http.MethodPost, spacesPath, `{"key":"eng","name":"開発部"}`)
		require.Equal(t, http.StatusCreated, spaceRes.Code, spaceRes.Body.String())
		var space kbSpaceResponse
		require.NoError(t, json.Unmarshal(spaceRes.Body.Bytes(), &space))

		// 4. 空のスペースに最初のページを作る（parentId 無し）。
		//    ここが通らないと、ページを 1 枚も持たないスペースは永久に空のままになる。
		pagesPath := spacesPath + "/" + space.ID + "/pages"
		pageRes := e.do(t, http.MethodPost, pagesPath, `{"title":"最初のページ"}`)
		require.Equal(t, http.StatusCreated, pageRes.Code, pageRes.Body.String())
		var page kbPageResponse
		require.NoError(t, json.Unmarshal(pageRes.Body.Bytes(), &page))
		assert.Nil(t, page.ParentID, "スペース直下（ルート）に作る")

		// 5. ツリーに現れる。
		tree := e.do(t, http.MethodGet, pagesPath, "")
		require.Equal(t, http.StatusOK, tree.Code)
		var body kbPageTreeRootResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &body))
		nodes := body.Pages
		require.Len(t, nodes, 1)
		assert.Equal(t, page.ID, nodes[0].Page.ID)
	})

	t.Run("所属していないワークスペースは一覧に漏れない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		kbInsertWorkspace(t, sqlDB, "rival")

		listed := env.as(alice).do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.NotContains(t, listed.Body.String(), "rival",
			"principals に行が無いワークスペースは 1 件も出さない")

		none := env.as(bob).do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, none.Code)
		assert.JSONEq(t, `[]`, none.Body.String())
	})

	t.Run("スペース作成はワークスペースのadminだけが通る", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		carol := kbInsertUser(t, sqlDB, "carol")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		env.joinWorkspace(t, bob, domain.GrantRoleEditor)

		spacesPath := "/api/v2/kb/workspaces/acme/spaces"
		assert.Equal(t, http.StatusForbidden,
			env.as(bob).do(t, http.MethodPost, spacesPath, `{"key":"ops","name":"運用部"}`).Code,
			"editor では作れない")
		assert.Equal(t, http.StatusNotFound,
			env.as(carol).do(t, http.MethodPost, spacesPath, `{"key":"ops","name":"運用部"}`).Code,
			"非メンバーにはワークスペースの実在を漏らさない")
		assert.Equal(t, http.StatusCreated,
			env.as(alice).do(t, http.MethodPost, spacesPath, `{"key":"ops","name":"運用部"}`).Code)
	})

	t.Run("スペースのeditorでも親ページで外されていればその下には作れない", func(t *testing.T) {
		// この経路が「スペースの権限」で判断されると、ページに張った deny が素通りする。
		// 親を指定した作成は必ずページ単位の判定（経路上の例外まで見る）を通ること。
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		parent := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")

		// bob はスペース全体では editor のまま、この親ページだけ閲覧を外される。
		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, parent, bobPrincipal.ID,
			domain.CapabilityView, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		facts, err := env.permissions.SpacePermissionFactsForUser(t.Context(), env.workspaceID, env.spaceID, bob)
		require.NoError(t, err)
		require.True(t, domain.ResolveScopePermission(*facts).CanEdit,
			"前提: スペース単位ではまだ編集できる（だからこそ経路の取り違えが穴になる）")

		denied := e.do(t, http.MethodPost, e.pagesPath(), `{"parentId":"`+parent+`","title":"子"}`)
		assert.Equal(t, http.StatusNotFound, denied.Code, denied.Body.String())

		// 同じ相手でも、親を指定しないルート作成はスペースの既定どおり通る。
		root := e.do(t, http.MethodPost, e.pagesPath(), `{"title":"自分のルート"}`)
		assert.Equal(t, http.StatusCreated, root.Code, root.Body.String())
	})

	t.Run("スペースの編集権限が無ければルートページを作れない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		viewer := kbInsertUser(t, sqlDB, "viewer")
		stranger := kbInsertUser(t, sqlDB, "stranger")
		env.joinWorkspace(t, viewer, domain.GrantRoleViewer)
		env.joinWorkspace(t, stranger, domain.GrantRoleAdmin)

		assert.Equal(t, http.StatusForbidden,
			env.as(viewer).do(t, http.MethodPost, env.as(viewer).pagesPath(), `{"title":"だめ"}`).Code,
			"閲覧はできる相手なので理由を返してよい")

		// 別テナントのスペース ID を自分のワークスペースの URL に混ぜても通らない。
		otherWS := kbInsertWorkspace(t, sqlDB, "rival")
		otherSpace := kbInsertSpace(t, sqlDB, otherWS, "eng")
		crossed := "/api/v2/kb/workspaces/acme/spaces/" + otherSpace + "/pages"
		assert.Equal(t, http.StatusNotFound,
			env.as(stranger).do(t, http.MethodPost, crossed, `{"title":"越境"}`).Code,
			"ワークスペースの admin でも別テナントのスペースには届かない")
	})

	t.Run("役割の無いメンバーにはスペースの実在を漏らさない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		nobody := kbInsertUser(t, sqlDB, "nobody")
		// 所属だけさせて grant は張らない（middleware は通るが中身は何も見えない）。
		_, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, nobody)
		require.NoError(t, err)

		e := env.as(nobody)
		hidden := e.do(t, http.MethodPost, e.pagesPath(), `{"title":"見えない"}`)
		missing := e.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/acme/spaces/"+kbNewUUID()+"/pages", `{"title":"無い"}`)

		assert.Equal(t, http.StatusNotFound, hidden.Code)
		assert.Equal(t, hidden.Code, missing.Code)
		assert.Equal(t, hidden.Body.String(), missing.Body.String(),
			"見えないスペースと存在しないスペースの応答は同じにする")
	})
}

// spacesPath はスペースの一覧 / 作成のパス。
func (e *kbEnv) spacesPath() string {
	return "/api/v2/kb/workspaces/" + e.slug + "/spaces"
}

// listSpaces はスペース一覧を叩いて key を並べて返す（順序も検証したいので key のまま）。
func (e *kbEnv) listSpaces(t *testing.T, userID uint64) (*httptest.ResponseRecorder, []string) {
	t.Helper()
	w := e.as(userID).do(t, http.MethodGet, e.spacesPath(), "")
	if w.Code != http.StatusOK {
		return w, nil
	}
	var got []kbSpaceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	keys := make([]string, 0, len(got))
	for _, s := range got {
		keys = append(keys, s.Key)
	}
	return w, keys
}

// TestKnowledgeBaseListSpacesAPI_Integration はスペース一覧を実 PostgreSQL で確かめる。
//
// この口はサイドバーの入口で、返す中身がそのまま「誰に何を見せるか」になる。
// 権限のふるいは domain 側にあるが、**そこへ渡す事実を集めるのは SQL** なので、
// 事実の集め方（どの grant が届くか）は本物の DB でしか確かめられない。
func TestKnowledgeBaseListSpacesAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("役割の届いているスペースだけがkey順で返る", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		ops := kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		kbInsertSpace(t, sqlDB, env.workspaceID, "hr")

		// 所属だけさせて、ワークスペース全体の grant は張らない。
		// これを張ると全スペースに届いてしまい、ふるいが効いているか分からなくなる。
		bob := kbInsertUser(t, sqlDB, "bob")
		bobPrincipal, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, bob)
		require.NoError(t, err)
		_, err = env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, ops, bobPrincipal.ID, domain.GrantRoleViewer,
		)
		require.NoError(t, err)

		w, keys := env.listSpaces(t, bob)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, []string{"ops"}, keys, "grant が届いているスペースだけ")
		assert.NotContains(t, w.Body.String(), "hr", "権限の無いスペースは名前も漏らさない")
		assert.NotContains(t, w.Body.String(), "eng")
	})

	t.Run("役割ごとに見え方が変わらない", func(t *testing.T) {
		// viewer から admin まで、閲覧できる役割ならどれでも一覧に出る
		// （出す / 出さないの境目は「役割が 1 つも無いか」であって役割の強さではない）。
		for _, role := range []domain.GrantRole{
			domain.GrantRoleViewer, domain.GrantRoleCommenter,
			domain.GrantRoleEditor, domain.GrantRoleAdmin,
		} {
			t.Run(string(role), func(t *testing.T) {
				env := newKbEnv(t, sqlDB, "acme")
				kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
				user := kbInsertUser(t, sqlDB, "u")
				principal, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, user)
				require.NoError(t, err)
				_, err = env.permissions.UpsertSpaceGrant(
					t.Context(), env.workspaceID, env.spaceID, principal.ID, role,
				)
				require.NoError(t, err)

				_, keys := env.listSpaces(t, user)

				assert.Equal(t, []string{"eng"}, keys)
			})
		}
	})

	t.Run("ワークスペース全体のadminには全スペースが見える", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)

		_, keys := env.listSpaces(t, alice)

		assert.Equal(t, []string{"eng", "ops"}, keys, "ワークスペースの grant は配下の全スペースへ届く")
	})

	t.Run("グループ経由の役割も届く", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		carol := kbInsertUser(t, sqlDB, "carol")
		carolPrincipal, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, carol)
		require.NoError(t, err)
		group, err := env.permissions.CreateGroupPrincipal(t.Context(), env.workspaceID, "開発チーム")
		require.NoError(t, err)
		require.NoError(t, env.permissions.AddGroupMember(
			t.Context(), env.workspaceID, group.ID, carolPrincipal.ID,
		))
		_, err = env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, env.spaceID, group.ID, domain.GrantRoleEditor,
		)
		require.NoError(t, err)

		_, keys := env.listSpaces(t, carol)

		assert.Equal(t, []string{"eng"}, keys, "所属グループ宛ての grant も自分に届く")
	})

	t.Run("スペース全員宛ての役割はそのスペースにだけ効く", func(t *testing.T) {
		// kind='space_all' の主体はスペース 1 つに紐づく。これを「自分」に畳んで
		// 全スペースへ効かせると、1 つのスペースの公開設定がテナント全体に波及する。
		env := newKbEnv(t, sqlDB, "acme")
		ops := kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		dave := kbInsertUser(t, sqlDB, "dave")
		_, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, dave)
		require.NoError(t, err)
		everyone, err := env.permissions.EnsureSpaceEveryonePrincipal(t.Context(), env.workspaceID, ops)
		require.NoError(t, err)
		_, err = env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, ops, everyone.ID, domain.GrantRoleViewer,
		)
		require.NoError(t, err)

		w, keys := env.listSpaces(t, dave)

		assert.Equal(t, []string{"ops"}, keys, "全員宛ての grant を張ったスペースだけ")
		assert.NotContains(t, w.Body.String(), "eng", "別スペースには波及しない")
	})

	t.Run("非メンバーはスペース全員宛ての役割でも見えない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		everyone, err := env.permissions.EnsureSpaceEveryonePrincipal(t.Context(), env.workspaceID, env.spaceID)
		require.NoError(t, err)
		_, err = env.permissions.UpsertSpaceGrant(
			t.Context(), env.workspaceID, env.spaceID, everyone.ID, domain.GrantRoleViewer,
		)
		require.NoError(t, err)
		stranger := kbInsertUser(t, sqlDB, "stranger")

		w, _ := env.listSpaces(t, stranger)

		assert.Equal(t, http.StatusNotFound, w.Code,
			"所属していない相手は middleware で止まる（「全員」に含まれない）")
	})

	t.Run("スペースが0件でも空配列", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		_, err := sqlDB.Exec(`DELETE FROM spaces WHERE workspace_id = $1`, env.workspaceID)
		require.NoError(t, err)
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)

		w, _ := env.listSpaces(t, alice)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `[]`, w.Body.String(), "null ではなく空配列")
	})

	t.Run("役割が1件も無いメンバーには空配列", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		nobody := kbInsertUser(t, sqlDB, "nobody")
		_, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, nobody)
		require.NoError(t, err)

		w, _ := env.listSpaces(t, nobody)

		require.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `[]`, w.Body.String(),
			"所属しているだけでは中身は見えない（スペースの実在も漏らさない）")
	})

	t.Run("存在しないワークスペースと権限の無いワークスペースの応答が同じ", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		kbInsertWorkspace(t, sqlDB, "rival")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		e := env.as(alice)

		unknown := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/no-such-workspace/spaces", "")
		foreign := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/rival/spaces", "")

		assert.Equal(t, http.StatusNotFound, unknown.Code)
		assert.Equal(t, unknown.Code, foreign.Code)
		assert.Equal(t, unknown.Body.String(), foreign.Body.String(),
			"slug の総当たりで他社テナントの実在が分からないこと")
	})

	t.Run("別テナントのスペースは混ざらない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		rival := kbInsertWorkspace(t, sqlDB, "rival")
		kbInsertSpace(t, sqlDB, rival, "secret")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)

		w, keys := env.listSpaces(t, alice)

		assert.Equal(t, []string{"eng"}, keys)
		assert.NotContains(t, w.Body.String(), "secret")
	})
}

// TestKnowledgeBasePrivateSpaceAPI_Integration はプライベートスペースが「スペース単位で
// 付与された相手」以外へ**どの読み取り経路からも**漏れないことを実 PostgreSQL で確かめる。
//
// ワークスペース全体の grant を見るクエリは 8 本あり、1 本でもふるい忘れると
// 「一覧には出ないのに URL 直叩きで読める」私室ができる。ここでは admin（ワークスペース
// 既定では最強の相手）を観測者にして、一覧・木・1 枚・検索の全部で見えないことを固定する。
func TestKnowledgeBasePrivateSpaceAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	env := newKbEnv(t, sqlDB, "acme")
	alice := kbInsertUser(t, sqlDB, "alice") // ワークスペースの admin（それでも見えない）
	bob := kbInsertUser(t, sqlDB, "bob")     // editor・プライベートスペースの作成者
	env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
	env.joinWorkspace(t, bob, domain.GrantRoleEditor)

	// bob（admin ではない）がプライベートスペースを作れる。
	spacesPath := "/api/v2/kb/workspaces/acme/spaces"
	created := env.as(bob).do(t, http.MethodPost, spacesPath,
		`{"name":"bob の下書き","visibility":"private"}`)
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	var space kbSpaceResponse
	require.NoError(t, json.Unmarshal(created.Body.Bytes(), &space))
	require.Equal(t, "private", space.Visibility)

	// 中にページも作れる（作成時に張られる space_grant(admin) が唯一の入口。
	// これが無ければワークスペース既定は届かず、作った本人にも見えない）。
	pagesPath := spacesPath + "/" + space.ID + "/pages"
	pageRes := env.as(bob).do(t, http.MethodPost, pagesPath, `{"title":"秘密のメモ"}`)
	require.Equal(t, http.StatusCreated, pageRes.Code, pageRes.Body.String())
	var page kbPageResponse
	require.NoError(t, json.Unmarshal(pageRes.Body.Bytes(), &page))

	t.Run("作成者にはすべての経路で見える", func(t *testing.T) {
		e := env.as(bob)
		listed := e.do(t, http.MethodGet, spacesPath, "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.Contains(t, listed.Body.String(), space.ID, "一覧に出る")

		tree := e.do(t, http.MethodGet, pagesPath, "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.Contains(t, tree.Body.String(), page.ID, "木に出る")

		got := e.do(t, http.MethodGet, env.pagePath(page.ID), "")
		assert.Equal(t, http.StatusOK, got.Code, "本文を開ける")

		search := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/search?q=秘密", "")
		require.Equal(t, http.StatusOK, search.Code)
		assert.Contains(t, search.Body.String(), page.ID, "検索に出る")
	})

	t.Run("ワークスペースのadminでもスペース単位の付与が無ければ何も見えない", func(t *testing.T) {
		e := env.as(alice)
		listed := e.do(t, http.MethodGet, spacesPath, "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.NotContains(t, listed.Body.String(), space.ID, "一覧に行ごと出さない")

		// スペース ID を知っていても木は空。404 にしないのは Tree の設計
		// （「無いスペース」と「見えないスペース」を撃ち分けると ID の総当たりで
		// 実在が分かるため、どちらも 200 の空配列）。中身が漏れないことが本題。
		tree := e.do(t, http.MethodGet, pagesPath, "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.NotContains(t, tree.Body.String(), page.ID, "木にページが載らない")
		assert.Contains(t, tree.Body.String(), `"pages":[]`, "1 件も出ない")

		// ページ ID を知っていても 1 枚を開けない。
		got := e.do(t, http.MethodGet, env.pagePath(page.ID), "")
		assert.Equal(t, http.StatusNotFound, got.Code, "URL 直叩きも 404")

		search := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/search?q=秘密", "")
		require.Equal(t, http.StatusOK, search.Code)
		assert.NotContains(t, search.Body.String(), page.ID, "検索にも出ない")
	})

	t.Run("チームスペースは今までどおり全員に見える", func(t *testing.T) {
		// 回帰の確認: visibility のふるいを足したことで、既定の 'workspace' の
		// スペースまで見えなくなっていないこと（env の既定スペース eng を使う）。
		e := env.as(alice)
		listed := e.do(t, http.MethodGet, spacesPath, "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.Contains(t, listed.Body.String(), env.spaceID)
	})
}

// TestKnowledgeBaseCompanyMembership_Integration は「同じ会社の人はチームスペースを見られる」
// を実 PostgreSQL で確かめる。
//
// 会社ごとのワークスペースは起動時のバックフィルが用意し users.workspace_id へ写すが、
// ノートの所属（principals の行）は作成者にしか無かったため、同じ会社の他の
// メンバーには一覧にも出ず URL も 404 だった。ここでその経路を固定する。
func TestKnowledgeBaseCompanyMembership_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	env := newKbEnv(t, sqlDB, "acme")
	alice := kbInsertUser(t, sqlDB, "alice") // ワークスペースを作った人
	bob := kbInsertUser(t, sqlDB, "bob")     // 同じ会社の別の人（principals の行は無い）
	carol := kbInsertUser(t, sqlDB, "carol") // 別の会社の人
	env.joinWorkspace(t, alice, domain.GrantRoleAdmin)

	// 会社の紐づけ（本番の tenant_bridge が埋める列）を用意する。
	otherWorkspaceID := kbInsertWorkspace(t, sqlDB, "rival")
	for _, c := range []struct {
		user uint64
		ws   string
	}{{alice, env.workspaceID}, {bob, env.workspaceID}, {carol, otherWorkspaceID}} {
		_, err := sqlDB.Exec(`UPDATE users SET workspace_id = $1 WHERE id = $2`, c.ws, c.user)
		require.NoError(t, err)
	}

	// alice がチームスペースへページを 1 枚置く。
	pageRes := env.as(alice).do(t, http.MethodPost, env.pagesPath(), `{"title":"全社の議事録"}`)
	require.Equal(t, http.StatusCreated, pageRes.Code, pageRes.Body.String())
	var page kbPageResponse
	require.NoError(t, json.Unmarshal(pageRes.Body.Bytes(), &page))

	t.Run("同じ会社の人は一覧・木・本文まで届く（所属は自動で用意される）", func(t *testing.T) {
		e := env.as(bob)

		listed := e.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.Contains(t, listed.Body.String(), "acme", "会社のワークスペースが一覧に出る")

		spaces := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/spaces", "")
		require.Equal(t, http.StatusOK, spaces.Code)
		assert.Contains(t, spaces.Body.String(), env.spaceID, "チームスペースが見える")

		tree := e.do(t, http.MethodGet, env.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.Contains(t, tree.Body.String(), page.ID, "木にページが出る")

		got := e.do(t, http.MethodGet, env.pagePath(page.ID), "")
		assert.Equal(t, http.StatusOK, got.Code, "本文を開ける")
	})

	t.Run("既定は editor なので書ける（読むだけの人を作らない）", func(t *testing.T) {
		w := env.as(bob).do(t, http.MethodPatch, env.pagePath(page.ID), `{"title":"議事録（改訂）"}`)
		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("別の会社の人には見えないまま", func(t *testing.T) {
		e := env.as(carol)

		listed := e.do(t, http.MethodGet, "/api/v2/kb/workspaces", "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.NotContains(t, listed.Body.String(), `"slug":"acme"`, "他社のワークスペースは出ない")

		spaces := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/spaces", "")
		assert.Equal(t, http.StatusNotFound, spaces.Code, "URL を知っていても 404")
	})

	t.Run("プライベートスペースは会社の全員には見えない", func(t *testing.T) {
		// 会社の全員が入っても、プライベートは付与された人だけ（ワークスペース全体の
		// grant は private へ届かない）。この 2 つが両立して初めて節分けが意味を持つ。
		created := env.as(bob).do(t, http.MethodPost, "/api/v2/kb/workspaces/acme/spaces",
			`{"name":"bob の下書き","visibility":"private"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var private kbSpaceResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &private))

		listed := env.as(alice).do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/spaces", "")
		require.Equal(t, http.StatusOK, listed.Code)
		assert.NotContains(t, listed.Body.String(), private.ID,
			"会社の admin にも、他人のプライベートスペースは見えない")
	})
}
