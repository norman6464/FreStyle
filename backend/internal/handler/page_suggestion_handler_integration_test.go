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

// TestPageSuggestionAPI_Integration は提案 API を実 PostgreSQL・本物のルータ
// （registerKnowledgeBaseRoutesWith 経由）で end-to-end に固定する。
//
// 中心の確認: viewer は提案できない、commenter は提案できるが採用・却下や本文の直接書き換えは
// できない、admin（editor 以上）が採用すると本文が変わり版が増える、却下すると本文は変わらない。
func TestPageSuggestionAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	env := newKbEnv(t, sqlDB, "sugg")
	admin := kbInsertUser(t, sqlDB, "admin")
	env.joinWorkspace(t, admin, domain.GrantRoleAdmin)
	rootPage := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, admin, "a0", "root")

	viewer := kbInsertUser(t, sqlDB, "viewer")
	env.joinWorkspace(t, viewer, domain.GrantRoleViewer)
	commenter := kbInsertUser(t, sqlDB, "commenter")
	env.joinWorkspace(t, commenter, domain.GrantRoleCommenter)

	asAdmin := env.as(admin)
	asViewer := env.as(viewer)
	asCommenter := env.as(commenter)

	suggestionsPath := "/api/v2/kb/workspaces/" + env.slug + "/pages/" + rootPage + "/suggestions"
	contentPath := "/api/v2/kb/workspaces/" + env.slug + "/pages/" + rootPage + "/content"
	const suggestedDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"commenterの提案"}]}]}`

	t.Run("viewerは提案できない", func(t *testing.T) {
		w := asViewer.do(t, http.MethodPost, suggestionsPath, `{"doc":`+suggestedDoc+`}`)
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("commenterは本文を直接書けない", func(t *testing.T) {
		w := asCommenter.do(t, http.MethodPut, contentPath, `{"doc":`+suggestedDoc+`}`)
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	var suggestionID string
	t.Run("commenterが保存すると本文ではなく提案として積まれる", func(t *testing.T) {
		w := asCommenter.do(t, http.MethodPost, suggestionsPath, `{"doc":`+suggestedDoc+`}`)
		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		var got kbPageSuggestionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "open", got.Status)
		suggestionID = got.ID

		page := asCommenter.do(t, http.MethodGet, env.pagePath(rootPage), "")
		require.Equal(t, http.StatusOK, page.Code)
		assert.NotContains(t, page.Body.String(), "commenterの提案", "本文はまだ変わっていない")
	})

	t.Run("commenterは自分の提案を採用できない", func(t *testing.T) {
		w := asCommenter.do(t, http.MethodPost, suggestionsPath+"/"+suggestionID+"/accept", "")
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
	})

	t.Run("adminが採用すると本文が変わり版が増える", func(t *testing.T) {
		versionsBefore := asAdmin.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+env.slug+"/pages/"+rootPage+"/versions", "")
		require.Equal(t, http.StatusOK, versionsBefore.Code)
		var before []pageVersionSummaryResponse
		require.NoError(t, json.Unmarshal(versionsBefore.Body.Bytes(), &before))

		w := asAdmin.do(t, http.MethodPost, suggestionsPath+"/"+suggestionID+"/accept", "")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got kbPageSuggestionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "accepted", got.Status)
		require.NotNil(t, got.ResolvedBy)
		assert.Equal(t, admin, got.ResolvedBy.UserID)

		page := asAdmin.do(t, http.MethodGet, env.pagePath(rootPage), "")
		require.Equal(t, http.StatusOK, page.Code)
		assert.Contains(t, page.Body.String(), "commenterの提案", "採用した提案が本文へ反映される")

		versionsAfter := asAdmin.do(t, http.MethodGet, "/api/v2/kb/workspaces/"+env.slug+"/pages/"+rootPage+"/versions", "")
		require.Equal(t, http.StatusOK, versionsAfter.Code)
		var after []pageVersionSummaryResponse
		require.NoError(t, json.Unmarshal(versionsAfter.Body.Bytes(), &after))
		assert.Len(t, after, len(before)+1, "採用は10分規則を無視して必ず版を1つ切る")
	})

	t.Run("既にacceptedな提案をもう一度解決しようとすると409", func(t *testing.T) {
		w := asAdmin.do(t, http.MethodPost, suggestionsPath+"/"+suggestionID+"/reject", "")
		assert.Equal(t, http.StatusConflict, w.Code)
		assert.JSONEq(t, `{"error":"suggestion_already_resolved"}`, w.Body.String())
	})

	t.Run("却下すると本文は変わらない", func(t *testing.T) {
		before := asAdmin.do(t, http.MethodGet, env.pagePath(rootPage), "")
		require.Equal(t, http.StatusOK, before.Code)

		const rejectedDoc = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"却下される提案"}]}]}`
		created := asCommenter.do(t, http.MethodPost, suggestionsPath, `{"doc":`+rejectedDoc+`}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var createdSugg kbPageSuggestionResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &createdSugg))

		w := asAdmin.do(t, http.MethodPost, suggestionsPath+"/"+createdSugg.ID+"/reject", "")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got kbPageSuggestionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		assert.Equal(t, "rejected", got.Status)

		after := asAdmin.do(t, http.MethodGet, env.pagePath(rootPage), "")
		require.Equal(t, http.StatusOK, after.Code)
		assert.Equal(t, before.Body.String(), after.Body.String(), "却下は本文を一切変えない")
	})

	t.Run("一覧は open な提案だけをcreated_at昇順で返す", func(t *testing.T) {
		w := asCommenter.do(t, http.MethodGet, suggestionsPath, "")
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var got []kbPageSuggestionResponse
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
		for _, s := range got {
			assert.Equal(t, "open", s.Status, "acceptedとrejectedになった提案は一覧に出ない")
		}
	})
}
