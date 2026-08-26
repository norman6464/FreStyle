//go:build integration

package handler

import (
	"encoding/json"
	"net/http"
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
		var nodes []kbPageTreeResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &nodes))
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
