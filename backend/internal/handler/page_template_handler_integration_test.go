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

// TestPageTemplateAPI_Integration はページの雛形 API（FRESTYLE-435 段5）を実 PostgreSQL・
// 本物のルータ（registerKnowledgeBaseRoutesWith 経由）で end-to-end に固定する。
//
// 中心の確認は権限の非対称性: ワークスペース全体の editor 未満（viewer・commenter）は
// 雛形の作成・削除ができない（ワークスペース全体への CanEdit を要求するため）が、
// 一覧は所属者なら誰でも読め、「使用」（雛形からページを作る）は行き先のスペース／ページの
// 編集権限だけで判定される（ワークスペース全体の役割とは独立）。
func TestPageTemplateAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("editor未満はテンプレートの作成削除ができないが一覧・使用はできる", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "tpl-perm")
		admin := kbInsertUser(t, sqlDB, "admin")
		env.joinWorkspace(t, admin, domain.GrantRoleAdmin)
		rootPage := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, admin, "a0", "root")

		commenter := kbInsertUser(t, sqlDB, "commenter")
		principal := env.joinWorkspace(t, commenter, domain.GrantRoleCommenter)
		// このスペースだけ editor を追加で張る。「使用」はこの段の権限だけで判定されることを
		// 確かめるため、ワークスペース全体は commenter のままにしておく。
		env.grantSpace(t, env.spaceID, principal.ID, domain.GrantRoleEditor)

		asAdmin := env.as(admin)
		asCommenter := env.as(commenter)

		// 一覧はワークスペース所属者なら誰でも読める（ワークスペース全体の役割を問わない）。
		listW := asCommenter.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+env.slug+"/templates", "")
		require.Equal(t, http.StatusOK, listW.Code, listW.Body.String())
		assert.JSONEq(t, `[]`, listW.Body.String())
		t.Logf("GET .../templates (0件) => %d %s", listW.Code, listW.Body.String())

		// 雛形の作成（保存）はワークスペース全体への CanEdit が要る。commenter はスペース単位の
		// editor は持っているが、ワークスペース全体では commenter のままなので 403。
		deniedCreate := asCommenter.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+rootPage+"/templates", `{"name":"テンプレ"}`)
		assert.Equal(t, http.StatusForbidden, deniedCreate.Code, deniedCreate.Body.String())
		t.Logf("POST .../pages/%s/templates (commenter) => %d %s", rootPage, deniedCreate.Code, deniedCreate.Body.String())

		// admin が代わりに雛形を作る。
		created := asAdmin.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/"+env.slug+"/pages/"+rootPage+"/templates", `{"name":"テンプレ"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		t.Logf("POST .../pages/%s/templates (admin) => %d %s", rootPage, created.Code, created.Body.String())
		var tpl kbPageTemplateResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &tpl))

		// 雛形の削除もワークスペース全体への CanEdit が要る。commenter は 403。
		deniedDelete := asCommenter.do(t, http.MethodDelete,
			"/api/v2/kb/workspaces/"+env.slug+"/templates/"+tpl.ID, "")
		assert.Equal(t, http.StatusForbidden, deniedDelete.Code, deniedDelete.Body.String())
		t.Logf("DELETE .../templates/%s (commenter) => %d %s", tpl.ID, deniedDelete.Code, deniedDelete.Body.String())

		// 使用（雛形からページを作る）は行き先のスペースの編集権限だけで判定される。
		// commenter はこのスペースに限り editor を持っているので通る
		// （ワークスペース全体の役割とは独立していることの確認）。
		useW := asCommenter.do(t, http.MethodPost,
			"/api/v2/kb/workspaces/"+env.slug+"/spaces/"+env.spaceID+"/pages/from-template",
			`{"templateId":"`+tpl.ID+`","title":"雛形から作成"}`)
		require.Equal(t, http.StatusCreated, useW.Code, useW.Body.String())
		t.Logf("POST .../spaces/%s/pages/from-template (commenter) => %d %s", env.spaceID, useW.Code, useW.Body.String())

		// admin はワークスペース全体の CanEdit を持つので、雛形の削除もできる。
		okDelete := asAdmin.do(t, http.MethodDelete,
			"/api/v2/kb/workspaces/"+env.slug+"/templates/"+tpl.ID, "")
		assert.Equal(t, http.StatusNoContent, okDelete.Code, okDelete.Body.String())
		t.Logf("DELETE .../templates/%s (admin) => %d", tpl.ID, okDelete.Code)
	})
}
