//go:build integration

package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKnowledgeBaseSearchAndBacklinksAPI_Integration は本文検索・逆リンクの
// 入口を実 PostgreSQL・本番と同じ配線（registerKnowledgeBaseRoutesWith
// 経由）で確かめる。GET .../search と GET .../pages/:pageId/backlinks を実際に叩き、
// 応答の形（matchField / excerpt / matchStart / matchLen、逆リンクの一覧）を固定する。
func TestKnowledgeBaseSearchAndBacklinksAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	env := newKbEnv(t, sqlDB, "acme")
	alice := kbInsertUser(t, sqlDB, "alice")
	env.joinWorkspace(t, alice, domain.GrantRoleEditor)
	e := env.as(alice)

	// 参照先ページ（逆リンクの対象）。
	targetCreated := e.do(t, http.MethodPost, e.pagesPath(), `{"title":"参照先ページ"}`)
	require.Equal(t, http.StatusCreated, targetCreated.Code, targetCreated.Body.String())
	var target kbPageResponse
	require.NoError(t, json.Unmarshal(targetCreated.Body.Bytes(), &target))

	// 題名一致の例（"Docker" を題名に含む）。
	titleMatchCreated := e.do(t, http.MethodPost, e.pagesPath(), `{"title":"Docker 入門"}`)
	require.Equal(t, http.StatusCreated, titleMatchCreated.Code, titleMatchCreated.Body.String())
	var titleMatchPage kbPageResponse
	require.NoError(t, json.Unmarshal(titleMatchCreated.Body.Bytes(), &titleMatchPage))

	// 本文一致の例（題名には "Docker" を含まないが、本文に含む。加えて参照先ページへの
	// pageRef を持たせ、逆リンクの参照元にもなる）。
	bodyMatchCreated := e.do(t, http.MethodPost, e.pagesPath(), `{"title":"使い方ガイド"}`)
	require.Equal(t, http.StatusCreated, bodyMatchCreated.Code, bodyMatchCreated.Body.String())
	var bodyMatchPage kbPageResponse
	require.NoError(t, json.Unmarshal(bodyMatchCreated.Body.Bytes(), &bodyMatchPage))

	saveDoc := `{"doc":{"type":"doc","content":[{"type":"paragraph","content":[
		{"type":"text","text":"この段落には Docker の使い方が書いてある"},
		{"type":"pageRef","attrs":{"pageId":"` + target.ID + `"}}
	]}]}}`
	saved := e.do(t, http.MethodPut, e.pagePath(bodyMatchPage.ID)+"/content", saveDoc)
	require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

	t.Run("検索: 題名一致と本文一致の両方を正しく返す", func(t *testing.T) {
		got := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/acme/search?q=docker", "")
		require.Equal(t, http.StatusOK, got.Code, got.Body.String())
		t.Logf("GET /api/v2/kb/workspaces/acme/search?q=docker => %s", got.Body.String())

		var results []kbSearchPageResponse
		require.NoError(t, json.Unmarshal(got.Body.Bytes(), &results))
		require.Len(t, results, 2, "題名一致・本文一致それぞれ1件ずつ")

		byID := map[string]kbSearchPageResponse{}
		for _, r := range results {
			byID[r.ID] = r
		}

		titleHit, ok := byID[titleMatchPage.ID]
		require.True(t, ok, "題名一致のページが返る")
		assert.Equal(t, "title", titleHit.MatchField)
		assert.Empty(t, titleHit.Excerpt, "題名一致は抜粋を出さない")

		bodyHit, ok := byID[bodyMatchPage.ID]
		require.True(t, ok, "本文一致のページが返る")
		assert.Equal(t, "body", bodyHit.MatchField)
		assert.Contains(t, bodyHit.Excerpt, "Docker", "本文一致は抜粋を返す")
		excerptRunes := []rune(bodyHit.Excerpt)
		require.LessOrEqual(t, bodyHit.MatchStart+bodyHit.MatchLen, len(excerptRunes),
			"matchStart/matchLen は excerpt の rune 範囲に収まる")
		assert.Equal(t, "Docker",
			string(excerptRunes[bodyHit.MatchStart:bodyHit.MatchStart+bodyHit.MatchLen]),
			"matchStart/matchLen はexcerptの中でヒットそのものを指す")
	})

	t.Run("逆リンク: このページを参照しているページの一覧を返す", func(t *testing.T) {
		got := e.do(t, http.MethodGet, e.pagePath(target.ID)+"/backlinks", "")
		require.Equal(t, http.StatusOK, got.Code, got.Body.String())
		t.Logf("GET %s/backlinks => %s", e.pagePath(target.ID), got.Body.String())

		var backlinks []kbPageResponse
		require.NoError(t, json.Unmarshal(got.Body.Bytes(), &backlinks))
		require.Len(t, backlinks, 1)
		assert.Equal(t, bodyMatchPage.ID, backlinks[0].ID)
	})

	t.Run("逆リンク: 参照されていないページは空配列", func(t *testing.T) {
		got := e.do(t, http.MethodGet, e.pagePath(titleMatchPage.ID)+"/backlinks", "")
		require.Equal(t, http.StatusOK, got.Code)
		assert.JSONEq(t, `[]`, got.Body.String())
	})

	t.Run("逆リンク: 見る権限が無いページは404", func(t *testing.T) {
		otherWS := newKbEnv(t, sqlDB, "rival")
		bob := kbInsertUser(t, sqlDB, "bob")
		otherWS.joinWorkspace(t, bob, domain.GrantRoleEditor)
		// bob は rival ワークスペースの所属者で、acme の target ページは見えない。
		got := otherWS.as(bob).do(t, http.MethodGet,
			"/api/v2/kb/workspaces/rival/pages/"+target.ID+"/backlinks", "")
		assert.Equal(t, http.StatusNotFound, got.Code, "テナントが違うページは404（存在を教えない）")
	})
}
