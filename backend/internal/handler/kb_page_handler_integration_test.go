//go:build integration

package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// kbIntegrationTables は結合テストが触るナレッジ基盤のテーブル（TRUNCATE 対象）。
// users は他の結合テストと共有するため消さず、毎回一意なアドレスで足す。
var kbIntegrationTables = []string{
	"share_links", "page_restrictions", "space_grants", "workspace_grants",
	"principal_members", "principals",
	"blocks", "page_paths", "page_snapshots", "pages", "spaces", "workspaces",
}

// kbEnv は本物の PostgreSQL・本物の repository・本番と同じルートで組んだ検証環境。
type kbEnv struct {
	pages       repository.KnowledgeBaseRepository
	permissions repository.KnowledgeBasePermissionRepository
	workspaceID string
	slug        string
	spaceID     string
	router      *gin.Engine
}

func newKbEnv(t *testing.T, gormDB *gorm.DB, sqlDB *sql.DB, slug string) *kbEnv {
	t.Helper()
	testsupport.TruncateAll(t, gormDB, kbIntegrationTables...)

	env := &kbEnv{
		pages:       persistence.NewKnowledgeBaseRepository(sqlDB),
		permissions: persistence.NewKnowledgeBasePermissionRepository(sqlDB),
		slug:        slug,
	}
	env.workspaceID = kbInsertWorkspace(t, sqlDB, slug)
	env.spaceID = kbInsertSpace(t, sqlDB, env.workspaceID, "eng")
	return env
}

// as は userID を current user としたルータを組む（本番と同じ registerKnowledgeBaseRoutesWith 経由）。
func (e *kbEnv) as(userID uint64) *kbEnv {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v2")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, userID)
		c.Next()
	})
	registerKnowledgeBaseRoutesWith(g, e.pages, e.permissions)
	clone := *e
	clone.router = r
	return &clone
}

func (e *kbEnv) do(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func (e *kbEnv) pagesPath() string {
	return "/api/v2/kb/workspaces/" + e.slug + "/spaces/" + e.spaceID + "/pages"
}

func (e *kbEnv) pagePath(pageID string) string {
	return "/api/v2/kb/workspaces/" + e.slug + "/pages/" + pageID
}

func kbNewUUID() string { return uuid.Must(uuid.NewV7()).String() }

func kbInsertWorkspace(t *testing.T, db *sql.DB, slug string) string {
	t.Helper()
	id := kbNewUUID()
	_, err := db.Exec(`INSERT INTO workspaces (id, slug, name) VALUES ($1, $2, $3)`, id, slug, slug)
	require.NoError(t, err)
	return id
}

func kbInsertSpace(t *testing.T, db *sql.DB, workspaceID, key string) string {
	t.Helper()
	id := kbNewUUID()
	_, err := db.Exec(`INSERT INTO spaces (id, workspace_id, "key", name) VALUES ($1, $2, $3, $3)`,
		id, workspaceID, key)
	require.NoError(t, err)
	return id
}

// kbInsertUser は users に 1 行入れて id を返す。principals が users へ FK を持つため、
// 権限を伴う結合テストは実在するユーザーを前提にする。
//
// id はシーケンスに任せず MAX(id)+1 で採番する。ほかの結合テストが users を
// TRUNCATE ... RESTART IDENTITY したり、id を明示して INSERT したりするため、
// シーケンスが実データより手前に取り残されていることがある（そのまま nextval を使うと
// 既存 id と衝突する）。結合テストは testsupport 側の advisory lock で直列化されているので、
// MAX(id)+1 が競合することはない。
//
// 逆に、明示採番した行を残すと今度は他のテストの nextval とぶつかるので、
// テスト終了時に必ず消す（users は共有テーブルなので TRUNCATE しない）。
func kbInsertUser(t *testing.T, db *sql.DB, name string) uint64 {
	t.Helper()
	var id uint64
	require.NoError(t, db.QueryRow(
		`INSERT INTO users (id, email, name, role_id, is_active, created_at, updated_at)
		 VALUES ((SELECT COALESCE(MAX(id), 0) + 1 FROM users), $1, $2, 3, true, now(), now())
		 RETURNING id`,
		name+"+"+kbNewUUID()+"@example.test", name,
	).Scan(&id))
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM users WHERE id = $1`, id); err != nil {
			t.Errorf("テストユーザーの後始末に失敗: %v", err)
		}
	})
	return id
}

// kbInsertRootPage はスペース直下のページを直接入れる（HTTP からは親付きしか作れないため）。
func kbInsertRootPage(t *testing.T, db *sql.DB, workspaceID, spaceID string, createdBy uint64, position, title string) string {
	t.Helper()
	id := kbNewUUID()
	_, err := db.Exec(
		`INSERT INTO pages (id, workspace_id, space_id, parent_id, "position", title, created_by_user_id)
		 VALUES ($1, $2, $3, NULL, $4, $5, $6)`,
		id, workspaceID, spaceID, position, title, createdBy,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO page_paths (workspace_id, page_id, ancestor_id, depth) VALUES ($1, $2, $2, 0)`,
		workspaceID, id,
	)
	require.NoError(t, err)
	return id
}

// joinWorkspace はユーザーをワークスペースに所属させ、既定の役割を与える。
func (e *kbEnv) joinWorkspace(t *testing.T, userID uint64, role domain.GrantRole) *domain.Principal {
	t.Helper()
	principal, err := e.permissions.EnsureUserPrincipal(t.Context(), e.workspaceID, userID)
	require.NoError(t, err)
	_, err = e.permissions.UpsertWorkspaceGrant(t.Context(), e.workspaceID, principal.ID, role)
	require.NoError(t, err)
	return principal
}

func TestKnowledgeBasePageAPI_Integration(t *testing.T) {
	gormDB := testsupport.OpenTestDB(t)
	sqlDB, err := gormDB.DB()
	require.NoError(t, err)

	t.Run("編集者はページを作って本文を保存し取得できる", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(alice)

		created := e.do(t, http.MethodPost, e.pagesPath(),
			`{"parentId":"`+root+`","title":"設計メモ"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var page kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &page))

		saved := e.do(t, http.MethodPut, e.pagePath(page.ID)+"/content",
			`{"doc":{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"本文"}]}]}}`)
		require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())

		got := e.do(t, http.MethodGet, e.pagePath(page.ID), "")
		require.Equal(t, http.StatusOK, got.Code)
		var doc kbPageDocResponse
		require.NoError(t, json.Unmarshal(got.Body.Bytes(), &doc))
		assert.Contains(t, string(doc.Doc), "本文")

		tree := e.do(t, http.MethodGet, e.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		var nodes []kbPageTreeResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &nodes))
		require.Len(t, nodes, 1)
		require.Len(t, nodes[0].Children, 1)
		assert.Equal(t, page.ID, nodes[0].Children[0].Page.ID)
	})

	t.Run("閲覧だけの役割は書き込みが403で読み取りは通る", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleEditor)
		env.joinWorkspace(t, bob, domain.GrantRoleViewer)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(bob)

		assert.Equal(t, http.StatusOK, e.do(t, http.MethodGet, e.pagePath(root), "").Code)
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPatch, e.pagePath(root), `{"title":"改訂"}`).Code)
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPost, e.pagePath(root)+"/archive", "").Code)
	})

	t.Run("所属していないワークスペースは404", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")

		// 同じ DB に別テナントを作り、alice はそちらには所属させない。
		otherWS := kbInsertWorkspace(t, sqlDB, "rival")
		otherSpace := kbInsertSpace(t, sqlDB, otherWS, "eng")
		e := env.as(alice)

		w := e.do(t, http.MethodGet, "/api/v2/kb/workspaces/rival/pages/"+root, "")
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.JSONEq(t, `{"error":"not_found"}`, w.Body.String())

		w = e.do(t, http.MethodGet, "/api/v2/kb/workspaces/rival/spaces/"+otherSpace+"/pages", "")
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ページの例外で隠した親の子はツリーに現れない", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)

		secret := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "秘密")
		open := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a5", "公開")

		// 秘密の下に子を作る（alice の権限で）。
		adminEnv := env.as(alice)
		created := adminEnv.do(t, http.MethodPost, adminEnv.pagesPath(),
			`{"parentId":"`+secret+`","title":"秘密の子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))

		// bob だけ「秘密」の閲覧を外す。子には例外を張らない（継承で消えるかを見る）。
		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, secret, bobPrincipal.ID,
			domain.CapabilityView, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		tree := e.do(t, http.MethodGet, e.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		var nodes []kbPageTreeResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &nodes))
		require.Len(t, nodes, 1, "隠した親も、その子も根に浮かない")
		assert.Equal(t, open, nodes[0].Page.ID)

		// 直リンクでも開けない（子は親の例外を継承する）。
		assert.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(secret), "").Code)
		assert.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(child.ID), "").Code)
	})

	t.Run("存在しないページと隠したページの応答が同じ", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		secret := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "秘密")
		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, secret, bobPrincipal.ID,
			domain.CapabilityView, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		hidden := e.do(t, http.MethodGet, e.pagePath(secret), "")
		missing := e.do(t, http.MethodGet, e.pagePath(kbNewUUID()), "")

		assert.Equal(t, http.StatusNotFound, hidden.Code)
		assert.Equal(t, hidden.Code, missing.Code)
		assert.Equal(t, hidden.Body.String(), missing.Body.String())
	})

	t.Run("編集を外した相手は改名できないが閲覧はできる", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		page := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "共有")
		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, page, bobPrincipal.ID,
			domain.CapabilityEdit, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		assert.Equal(t, http.StatusOK, e.do(t, http.MethodGet, e.pagePath(page), "").Code)
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPatch, e.pagePath(page), `{"title":"改訂"}`).Code)
	})

	t.Run("移動先の親に編集権限が無ければ移せない", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		src := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "移動元")
		dest := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a5", "移動先")
		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, dest, bobPrincipal.ID,
			domain.CapabilityEdit, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		w := e.do(t, http.MethodPost, e.pagePath(src)+"/move", `{"parentId":"`+dest+`"}`)
		assert.Equal(t, http.StatusForbidden, w.Code, "移動元だけ編集できても移せない")

		// alice（管理者）なら移せる。
		admin := env.as(alice)
		ok := admin.do(t, http.MethodPost, admin.pagePath(src)+"/move", `{"parentId":"`+dest+`"}`)
		assert.Equal(t, http.StatusOK, ok.Code, ok.Body.String())
	})

	t.Run("アーカイブした子孫はツリーから消え復帰で戻る", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(alice)

		created := e.do(t, http.MethodPost, e.pagesPath(), `{"parentId":"`+root+`","title":"子"}`)
		require.Equal(t, http.StatusCreated, created.Code)

		require.Equal(t, http.StatusNoContent, e.do(t, http.MethodPost, e.pagePath(root)+"/archive", "").Code)
		tree := e.do(t, http.MethodGet, e.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.JSONEq(t, `[]`, tree.Body.String())

		require.Equal(t, http.StatusOK, e.do(t, http.MethodPost, e.pagePath(root)+"/unarchive", "").Code)
		tree = e.do(t, http.MethodGet, e.pagesPath(), "")
		var nodes []kbPageTreeResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &nodes))
		require.Len(t, nodes, 1)
		assert.Len(t, nodes[0].Children, 1, "一緒にアーカイブした子も戻る")
	})

	t.Run("役割が無いメンバーは何も見えない", func(t *testing.T) {
		env := newKbEnv(t, gormDB, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		carol := kbInsertUser(t, sqlDB, "carol")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		// carol は所属だけ（grant なし）。
		_, err := env.permissions.EnsureUserPrincipal(t.Context(), env.workspaceID, carol)
		require.NoError(t, err)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")

		e := env.as(carol)
		assert.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(root), "").Code)
		tree := e.do(t, http.MethodGet, e.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.JSONEq(t, `[]`, tree.Body.String())
	})
}
