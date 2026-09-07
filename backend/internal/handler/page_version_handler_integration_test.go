//go:build integration

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pageVersionsPath / pageVersionPath / pageVersionRestorePath は版 API の URL 組み立て。
func (e *kbEnv) pageVersionsPath(pageID string) string {
	return e.pagePath(pageID) + "/versions"
}

func (e *kbEnv) pageVersionPath(pageID string, seq int64) string {
	return e.pagePath(pageID) + "/versions/" + strconv.FormatInt(seq, 10)
}

func (e *kbEnv) pageVersionRestorePath(pageID string, seq int64) string {
	return e.pageVersionPath(pageID, seq) + "/restore"
}

// TestPageVersionAPI_Integration は版 API（一覧・単体取得・作成・復元）を本物の
// PostgreSQL・本物のルータ（registerKnowledgeBaseRoutesWith 経由）で end-to-end に固定する。
func TestPageVersionAPI_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("本文保存で版ができ、一覧・単体取得・版を残す・復元が一通り通る", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(alice)

		// 1 回目の本文保存 = seq=1（版が1件も無い状態からの初回保存は必ず作られる）。
		docA := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"版A"}]}]}`
		saved := e.do(t, http.MethodPut, e.pagePath(root)+"/content", `{"doc":`+docA+`}`)
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

		// 10分規則を跨がせるため、直近の版の created_at を直接 SQL で 11 分前に見せかける
		// （time.Now() 自体はモックしない設計。実時刻のまま境界を再現する）。
		_, err := sqlDB.Exec(
			`UPDATE page_versions SET created_at = $1 WHERE workspace_id = $2 AND page_id = $3 AND seq = 1`,
			time.Now().Add(-11*time.Minute), env.workspaceID, root,
		)
		require.NoError(t, err)

		// 2 回目の本文保存 = seq=2。
		docB := `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"版B"}]}]}`
		saved = e.do(t, http.MethodPut, e.pagePath(root)+"/content", `{"doc":`+docB+`}`)
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

		// 「版を残す」= seq=3（10分規則を無視して必ず作られる。note が保存される）。
		created := e.do(t, http.MethodPost, e.pageVersionsPath(root), `{"note":"リリース直前の状態"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		t.Logf("POST .../versions (版を残す) response: %s", created.Body.String())
		var createdBody pageVersionDetailResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &createdBody))
		assert.Equal(t, int64(3), createdBody.Seq)
		require.NotNil(t, createdBody.Note)
		assert.Equal(t, "リリース直前の状態", *createdBody.Note)
		assert.Equal(t, alice, createdBody.Author.UserID)
		assert.Equal(t, "alice", createdBody.Author.Name)

		// 一覧は seq 降順で 3 件（本文保存 x2 + 版を残す x1）。
		list := e.do(t, http.MethodGet, e.pageVersionsPath(root), "")
		require.Equal(t, http.StatusOK, list.Code, list.Body.String())
		t.Logf("GET .../versions (一覧) response: %s", list.Body.String())
		var versions []pageVersionSummaryResponse
		require.NoError(t, json.Unmarshal(list.Body.Bytes(), &versions))
		require.Len(t, versions, 3)
		assert.Equal(t, []int64{3, 2, 1}, []int64{versions[0].Seq, versions[1].Seq, versions[2].Seq})

		// 単体取得（seq=1 = docA）は doc を含む。
		got := e.do(t, http.MethodGet, e.pageVersionPath(root, 1), "")
		require.Equal(t, http.StatusOK, got.Code, got.Body.String())
		t.Logf("GET .../versions/1 (単体取得) response: %s", got.Body.String())
		var gotBody pageVersionDetailResponse
		require.NoError(t, json.Unmarshal(got.Body.Bytes(), &gotBody))
		assert.Equal(t, int64(1), gotBody.Seq)
		assert.Contains(t, string(gotBody.Doc), "版A")

		// 復元（seq=1 = docA を書き戻す）は PUT .../content と同じ応答形。
		restored := e.do(t, http.MethodPost, e.pageVersionRestorePath(root, 1), "")
		require.Equal(t, http.StatusOK, restored.Code, restored.Body.String())
		t.Logf("POST .../versions/1/restore (復元) response: %s", restored.Body.String())
		var restoredBody kbPageContentResponse
		require.NoError(t, json.Unmarshal(restored.Body.Bytes(), &restoredBody))
		assert.Contains(t, string(restoredBody.Doc), "版A")
		require.NotNil(t, restoredBody.LastEditedBy)
		assert.Equal(t, alice, restoredBody.LastEditedBy.UserID)

		// 本文が実際に docA へ戻っていること（GetPageUseCase 経由の GET でも確認）。
		page := e.do(t, http.MethodGet, e.pagePath(root), "")
		require.Equal(t, http.StatusOK, page.Code)
		var pageDoc kbPageDocResponse
		require.NoError(t, json.Unmarshal(page.Body.Bytes(), &pageDoc))
		assert.Contains(t, string(pageDoc.Doc), "版A")

		// 復元自体で版が 1 つ増えている（seq=4）。
		listAfterRestore := e.do(t, http.MethodGet, e.pageVersionsPath(root), "")
		require.Equal(t, http.StatusOK, listAfterRestore.Code)
		var versionsAfter []pageVersionSummaryResponse
		require.NoError(t, json.Unmarshal(listAfterRestore.Body.Bytes(), &versionsAfter))
		require.Len(t, versionsAfter, 4, "復元自体も版になる")
		assert.Equal(t, int64(4), versionsAfter[0].Seq)
	})

	t.Run("閲覧だけの役割は一覧取得できるが作成と復元は403", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		env.joinWorkspace(t, bob, domain.GrantRoleViewer)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		asAlice := env.as(alice)
		saved := asAlice.do(t, http.MethodPut, asAlice.pagePath(root)+"/content",
			`{"doc":{"type":"doc","content":[]}}`)
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

		asBob := env.as(bob)
		assert.Equal(t, http.StatusOK, asBob.do(t, http.MethodGet, asBob.pageVersionsPath(root), "").Code)
		assert.Equal(t, http.StatusOK, asBob.do(t, http.MethodGet, asBob.pageVersionPath(root, 1), "").Code)
		assert.Equal(t, http.StatusForbidden,
			asBob.do(t, http.MethodPost, asBob.pageVersionsPath(root), "{}").Code)
		assert.Equal(t, http.StatusForbidden,
			asBob.do(t, http.MethodPost, asBob.pageVersionRestorePath(root, 1), "").Code)
	})

	t.Run("不正なnoteは400invalid_version_note", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(alice)

		note := make([]byte, domain.PageVersionNoteMaxBytes+1)
		for i := range note {
			note[i] = 'a'
		}
		noteJSON, err := json.Marshal(string(note))
		require.NoError(t, err)

		w := e.do(t, http.MethodPost, e.pageVersionsPath(root), `{"note":`+string(noteJSON)+`}`)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"invalid_version_note"}`, w.Body.String())
	})

	t.Run("他テナント他ページの版は番号だけ変えても覗けない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		rootA := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "pageA")
		rootB := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a1", "pageB")
		e := env.as(alice)

		saved := e.do(t, http.MethodPut, e.pagePath(rootA)+"/content", `{"doc":{"type":"doc","content":[]}}`)
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

		// pageA の seq=1 を pageB の URL で覗く。
		w := e.do(t, http.MethodGet, e.pageVersionPath(rootB, 1), "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())

		// 数値でない seq も 404。
		w2 := e.do(t, http.MethodGet, e.pagePath(rootA)+"/versions/not-a-number", "")
		assert.Equal(t, http.StatusNotFound, w2.Code)
	})
}
