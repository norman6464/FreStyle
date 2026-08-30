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
)

// kbIntegrationTables は結合テストが触るノートのテーブル（TRUNCATE 対象）。
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
	provisioner repository.WorkspaceProvisioner
	users       repository.UserRepository
	workspaceID string
	slug        string
	spaceID     string
	router      *gin.Engine
}

func newKbEnv(t *testing.T, sqlDB *sql.DB, slug string) *kbEnv {
	t.Helper()
	testsupport.TruncateAll(t, sqlDB, kbIntegrationTables...)

	env := &kbEnv{
		pages:       persistence.NewKnowledgeBaseRepository(sqlDB),
		permissions: persistence.NewKnowledgeBasePermissionRepository(sqlDB),
		provisioner: persistence.NewWorkspaceProvisioner(sqlDB),
		users:       persistence.NewUserRepository(sqlDB),
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
	registerKnowledgeBaseRoutesWith(g, e.pages, e.permissions, e.provisioner, e.users, (&kbAuditRecorder{}).handler())
	// 認証不要のルート（共有リンクの検証）は current user を注入しない group に張る。
	// 本番の NewRouter と同じ位置関係にしないと「未認証でも通ること」を確かめられない。
	registerKnowledgeBasePublicRoutesWith(r.Group("/api/v2"), e.pages, e.permissions)
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

// kbInsertCompanyWithWorkspace は companies に 1 行入れ、workspace_id を明示して返す。
// 会社とワークスペースの 1 : 1 の紐付け（companies.workspace_id）を結合テストで
// 再現するための最小構成（本番は起動時のバックフィルが埋める）。
func kbInsertCompanyWithWorkspace(t *testing.T, db *sql.DB, name, workspaceID string) uint64 {
	t.Helper()
	var id uint64
	require.NoError(t, db.QueryRow(
		`INSERT INTO companies (id, name, is_active, workspace_id, created_at, updated_at)
		 VALUES ((SELECT COALESCE(MAX(id), 0) + 1 FROM companies), $1, true, $2, now(), now())
		 RETURNING id`,
		name, workspaceID,
	).Scan(&id))
	t.Cleanup(func() {
		if _, err := db.Exec(`DELETE FROM companies WHERE id = $1`, id); err != nil {
			t.Errorf("テスト会社の後始末に失敗: %v", err)
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
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("編集者はページを作って本文を保存し取得できる", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
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
		var body kbPageTreeRootResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &body))
		nodes := body.Pages
		require.Len(t, nodes, 1)
		require.Len(t, nodes[0].Children, 1)
		assert.Equal(t, page.ID, nodes[0].Children[0].Page.ID)
	})

	t.Run("閲覧だけの役割は書き込みが403で読み取りは通る", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
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
		env := newKbEnv(t, sqlDB, "acme")
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
		env := newKbEnv(t, sqlDB, "acme")
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
		var body kbPageTreeRootResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &body))
		nodes := body.Pages
		require.Len(t, nodes, 1, "隠した親も、その子も根に浮かない")
		assert.Equal(t, open, nodes[0].Page.ID)

		// 直リンクでも開けない（子は親の例外を継承する）。
		assert.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(secret), "").Code)
		assert.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(child.ID), "").Code)
	})

	t.Run("存在しないページと隠したページの応答が同じ", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
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
		env := newKbEnv(t, sqlDB, "acme")
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
		env := newKbEnv(t, sqlDB, "acme")
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
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "root")
		e := env.as(alice)

		created := e.do(t, http.MethodPost, e.pagesPath(), `{"parentId":"`+root+`","title":"子"}`)
		require.Equal(t, http.StatusCreated, created.Code)

		require.Equal(t, http.StatusNoContent, e.do(t, http.MethodPost, e.pagePath(root)+"/archive", "").Code)
		tree := e.do(t, http.MethodGet, e.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		assert.JSONEq(t, `{"pages":[],"hasHiddenChildren":false}`, tree.Body.String())

		require.Equal(t, http.StatusOK, e.do(t, http.MethodPost, e.pagePath(root)+"/unarchive", "").Code)
		tree = e.do(t, http.MethodGet, e.pagesPath(), "")
		var body kbPageTreeRootResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &body))
		nodes := body.Pages
		require.Len(t, nodes, 1)
		assert.Len(t, nodes[0].Children, 1, "一緒にアーカイブした子も戻る")
	})

	t.Run("スペース全員宛ての例外が残るサブツリーの別スペースへの移動は409", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		parent := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")
		e := env.as(alice)

		created := e.do(t, http.MethodPost, e.pagesPath(), `{"parentId":"`+parent+`","title":"子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))

		// 「このスペースの全員」宛ての例外。別スペースへ移ると行だけが残って評価されなくなる。
		//
		// mode を allow にしてあるのは、deny だと**動かす本人まで締め出される**ため。
		// space_all は「そのスペースの全員」なので、ワークスペースのメンバーは全員それを
		// 自分の主体として持つ（権限クエリの mine CTE）。deny を張ると alice 自身が
		// 子を見られなくなり、サブツリーの編集検査に先に引っかかって 403 になる
		// （それはそれで正しい応答だが、この test で見たいのは移動先スペースの検査）。
		// allow なら alice は許可リストに載っている側なので通り、行は space_all 宛てのまま残る。
		everyone, err := env.permissions.EnsureSpaceEveryonePrincipal(t.Context(), env.workspaceID, env.spaceID)
		require.NoError(t, err)
		_, err = env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, child.ID, everyone.ID,
			domain.CapabilityView, domain.RestrictionModeAllow,
		)
		require.NoError(t, err)

		otherSpace := kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		dest := kbInsertRootPage(t, sqlDB, env.workspaceID, otherSpace, alice, "a0", "移動先")

		w := e.do(t, http.MethodPost, e.pagePath(parent)+"/move", `{"parentId":"`+dest+`"}`)
		require.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"space_restriction_voided"}`, w.Body.String(),
			"正当な業務エラーであって DB 障害ではない（500 だとクライアントが再試行してよいと誤解する）")

		got := e.do(t, http.MethodGet, e.pagePath(parent), "")
		require.Equal(t, http.StatusOK, got.Code)
		var page kbPageDocResponse
		require.NoError(t, json.Unmarshal(got.Body.Bytes(), &page))
		assert.Equal(t, env.spaceID, page.Page.SpaceID, "移動はロールバックされている")
	})

	t.Run("見えない子を持つ親はアーカイブできない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		parent := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")

		adminEnv := env.as(alice)
		created := adminEnv.do(t, http.MethodPost, adminEnv.pagesPath(),
			`{"parentId":"`+parent+`","title":"見えない子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))

		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, child.ID, bobPrincipal.ID,
			domain.CapabilityView, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		require.Equal(t, http.StatusNotFound, e.do(t, http.MethodGet, e.pagePath(child.ID), "").Code,
			"bob には子が見えていない")

		w := e.do(t, http.MethodPost, e.pagePath(parent)+"/archive", "")
		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"subtree_forbidden"}`, w.Body.String())

		// 管理者のツリーからも消えていない（見えないページを黙って消せない）。
		tree := adminEnv.do(t, http.MethodGet, adminEnv.pagesPath(), "")
		require.Equal(t, http.StatusOK, tree.Code)
		var body kbPageTreeRootResponse
		require.NoError(t, json.Unmarshal(tree.Body.Bytes(), &body))
		nodes := body.Pages
		require.Len(t, nodes, 1)
		require.Len(t, nodes[0].Children, 1)
		assert.Equal(t, child.ID, nodes[0].Children[0].Page.ID)
	})

	t.Run("編集できない子を持つ親はアーカイブできず復帰もできない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		parent := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")

		adminEnv := env.as(alice)
		created := adminEnv.do(t, http.MethodPost, adminEnv.pagesPath(),
			`{"parentId":"`+parent+`","title":"読めるが書けない子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))

		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, child.ID, bobPrincipal.ID,
			domain.CapabilityEdit, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		require.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPatch, e.pagePath(child.ID), `{"title":"改訂"}`).Code,
			"子を直接改名すると 403")
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPost, e.pagePath(parent)+"/archive", "").Code,
			"親のアーカイブ経由でも同じ判定になる")

		// 管理者がアーカイブしたあと、bob は復帰もできない（片側だけ緩くしない）。
		require.Equal(t, http.StatusNoContent,
			adminEnv.do(t, http.MethodPost, adminEnv.pagePath(parent)+"/archive", "").Code)
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPost, e.pagePath(parent)+"/unarchive", "").Code)
		require.Equal(t, http.StatusOK,
			adminEnv.do(t, http.MethodPost, adminEnv.pagePath(parent)+"/unarchive", "").Code,
			"全部編集できる管理者は通る（アーカイブが常に失敗する締め方にはしない）")
	})

	t.Run("役割が無いメンバーは何も見えない", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
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
		assert.JSONEq(t, `{"pages":[],"hasHiddenChildren":false}`, tree.Body.String())
	})
}

// kbDumpTreeState はサブツリーの配置を決める列を丸ごと写し取る。
//
// 移動が書き換えるのは pages.parent_id / pages."position" / pages.space_id と
// closure（page_paths）の 4 つだけなので、この文字列が前後で一致すれば
// 「何も書き換わっていない」と言える。個別に assert を並べるより、
// 見落としが起きにくい（列が増えたらここへ足す）。
func kbDumpTreeState(t *testing.T, db *sql.DB, workspaceID string) string {
	t.Helper()
	var pages string
	require.NoError(t, db.QueryRow(
		`SELECT COALESCE(string_agg(
		     id::text || '|' || COALESCE(parent_id::text, '-') || '|' || "position" || '|' || space_id::text,
		     E'\n' ORDER BY id), '')
		 FROM pages WHERE workspace_id = $1`, workspaceID,
	).Scan(&pages))
	var paths string
	require.NoError(t, db.QueryRow(
		`SELECT COALESCE(string_agg(
		     page_id::text || '|' || ancestor_id::text || '|' || depth::text,
		     E'\n' ORDER BY page_id, depth), '')
		 FROM page_paths WHERE workspace_id = $1`, workspaceID,
	).Scan(&paths))
	return pages + "\n--\n" + paths
}

// TestKnowledgeBaseMovePermission_Integration は「移動は根 1 枚の権限しか見ていない」
// 穴が塞がっていることを実 PostgreSQL で確かめる。
//
// 移動はサブツリーごと動くので、子孫それぞれの祖先の並びが変わる。ページの例外は
// 経路の上から効くため、祖先が変われば子孫の実効権限も変わる。操作者から見えない
// 子孫の権限が本人の知らないうちに書き換わる、というのが塞ぐ相手。
//
// 判定はアーカイブ / 復帰と同じ（サブツリー全体を編集できなければ 403 subtree_forbidden）。
// closure まで含めて「断ったら何も書き換わらない」ことを見るので結合テストに置く
// （page_paths は fake が持っていない）。
func TestKnowledgeBaseMovePermission_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	// setup は 親(root) → 子(child) と、移動先(dest) を用意して bob の principal を返す。
	// bob はワークスペース全体では editor（root と dest は編集できる）。
	setup := func(t *testing.T, env *kbEnv, alice, bob uint64) (string, string, string, *domain.Principal) {
		t.Helper()
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		bobPrincipal := env.joinWorkspace(t, bob, domain.GrantRoleEditor)
		root := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")
		dest := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a2", "移動先")

		adminEnv := env.as(alice)
		created := adminEnv.do(t, http.MethodPost, adminEnv.pagesPath(),
			`{"parentId":"`+root+`","title":"子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))
		return root, child.ID, dest, bobPrincipal
	}

	cases := map[string]domain.Capability{
		"編集だけ外した子": domain.CapabilityEdit,
		"閲覧ごと外した子": domain.CapabilityView,
	}
	for name, capability := range cases {
		t.Run(name+"を持つ親は移動できず何も書き換わらない", func(t *testing.T) {
			env := newKbEnv(t, sqlDB, "acme")
			alice := kbInsertUser(t, sqlDB, "alice")
			bob := kbInsertUser(t, sqlDB, "bob")
			root, child, dest, bobPrincipal := setup(t, env, alice, bob)

			_, err := env.permissions.UpsertPageRestriction(
				t.Context(), env.workspaceID, child, bobPrincipal.ID,
				capability, domain.RestrictionModeDeny,
			)
			require.NoError(t, err)

			before := kbDumpTreeState(t, sqlDB, env.workspaceID)
			e := env.as(bob)
			w := e.do(t, http.MethodPost, e.pagePath(root)+"/move", `{"parentId":"`+dest+`"}`)

			assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			assert.JSONEq(t, `{"error":"subtree_forbidden"}`, w.Body.String())
			assert.Equal(t, before, kbDumpTreeState(t, sqlDB, env.workspaceID),
				"断ったなら parent_id / position / page_paths のどれも動かない")
		})
	}

	t.Run("子孫まで編集できるなら通ってclosureも張り替わる", func(t *testing.T) {
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		root, child, dest, _ := setup(t, env, alice, bob)

		e := env.as(bob)
		w := e.do(t, http.MethodPost, e.pagePath(root)+"/move", `{"parentId":"`+dest+`"}`)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		// 例外が 1 つも無いのが普通の状態なので、通常の移動まで止めない。
		var parentID string
		require.NoError(t, sqlDB.QueryRow(
			`SELECT parent_id::text FROM pages WHERE workspace_id = $1 AND id = $2`,
			env.workspaceID, root,
		).Scan(&parentID))
		assert.Equal(t, dest, parentID)

		// 子の祖先に移動先が加わる（＝ 継承の経路が変わる）。これがそのまま
		// 「見えない子孫の権限が変わる」の中身で、だから移動でも子孫を見る。
		var depth int
		require.NoError(t, sqlDB.QueryRow(
			`SELECT depth FROM page_paths WHERE workspace_id = $1 AND page_id = $2 AND ancestor_id = $3`,
			env.workspaceID, child, dest,
		).Scan(&depth))
		assert.Equal(t, 2, depth, "子から見て移動先は 2 段上の祖先になる")
	})

	t.Run("スペース全員宛てのdenyは動かす本人も締め出す", func(t *testing.T) {
		// space_all は「そのスペースの全員」なので、ワークスペースのメンバーは全員それを
		// 自分の主体として持つ。子に deny を張ると、張った admin 自身もその子を見られなくなり、
		// 親の移動はサブツリーの検査で止まる。
		//
		// これは意図した結果。見えない子孫を巻き込む移動を断るのがこの検査の役目で、
		// 「自分で張った例外だから自分は例外」という抜け道は作らない
		// （権限は誰が張ったかではなく、いま誰に何が届いているかだけで決まる）。
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		env.joinWorkspace(t, alice, domain.GrantRoleAdmin)
		parent := kbInsertRootPage(t, sqlDB, env.workspaceID, env.spaceID, alice, "a0", "親")
		e := env.as(alice)

		created := e.do(t, http.MethodPost, e.pagesPath(), `{"parentId":"`+parent+`","title":"子"}`)
		require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
		var child kbPageResponse
		require.NoError(t, json.Unmarshal(created.Body.Bytes(), &child))

		everyone, err := env.permissions.EnsureSpaceEveryonePrincipal(t.Context(), env.workspaceID, env.spaceID)
		require.NoError(t, err)
		_, err = env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, child.ID, everyone.ID,
			domain.CapabilityView, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		otherSpace := kbInsertSpace(t, sqlDB, env.workspaceID, "ops")
		dest := kbInsertRootPage(t, sqlDB, env.workspaceID, otherSpace, alice, "a0", "移動先")

		before := kbDumpTreeState(t, sqlDB, env.workspaceID)
		w := e.do(t, http.MethodPost, e.pagePath(parent)+"/move", `{"parentId":"`+dest+`"}`)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		assert.JSONEq(t, `{"error":"subtree_forbidden"}`, w.Body.String(),
			"移動先スペースの検査（409）より手前で断る")
		assert.Equal(t, before, kbDumpTreeState(t, sqlDB, env.workspaceID))
	})

	t.Run("同一スペース内の移動でも子孫を見る", func(t *testing.T) {
		// ErrPageMoveVoidsSpaceRestriction が塞いでいるのはスペースをまたぐ移動だけ。
		// 同一スペース内で親を付け替える移動には、子孫の権限を見る経路がこれしかない。
		env := newKbEnv(t, sqlDB, "acme")
		alice := kbInsertUser(t, sqlDB, "alice")
		bob := kbInsertUser(t, sqlDB, "bob")
		root, child, dest, bobPrincipal := setup(t, env, alice, bob)

		var spaceIDs int
		require.NoError(t, sqlDB.QueryRow(
			`SELECT count(DISTINCT space_id) FROM pages WHERE workspace_id = $1`,
			env.workspaceID,
		).Scan(&spaceIDs))
		require.Equal(t, 1, spaceIDs, "前提: 移動元も移動先も同じスペース")

		_, err := env.permissions.UpsertPageRestriction(
			t.Context(), env.workspaceID, child, bobPrincipal.ID,
			domain.CapabilityEdit, domain.RestrictionModeDeny,
		)
		require.NoError(t, err)

		e := env.as(bob)
		require.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPatch, e.pagePath(child), `{"title":"改訂"}`).Code,
			"子を直接改名すると 403")
		assert.Equal(t, http.StatusForbidden,
			e.do(t, http.MethodPost, e.pagePath(root)+"/move", `{"parentId":"`+dest+`"}`).Code,
			"親の移動経由でも同じ判定になる（アーカイブと揃えてある）")

		// 全部編集できる管理者は通る（移動が常に失敗する締め方にはしない）。
		adminEnv := env.as(alice)
		assert.Equal(t, http.StatusOK,
			adminEnv.do(t, http.MethodPost, adminEnv.pagePath(root)+"/move", `{"parentId":"`+dest+`"}`).Code)
	})
}
