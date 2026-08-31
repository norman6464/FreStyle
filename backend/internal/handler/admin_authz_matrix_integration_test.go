//go:build integration

package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
)

// 管理系エンドポイントの認可を、本番と同じ配線（RequireAdmin + 各 handler の検査）で
// 役割ごとに実測する。
//
// 画面側の出し分けをやめると、backend の検査だけが境界になる。「UI に出していないから
// 誰も叩かない」で成り立っていた経路があれば、そこが穴になる。実際に過去 2 度、handler の
// 検査漏れで trainee が管理データを読めている（RequireAdmin のコメント参照）ので、
// 読んで確かめるのではなく叩いて確かめる。

// authzActor は検証で使う「叩く人」。
type authzActor struct {
	name string
	user *domain.User
}

func authzWorkspace(n int) string {
	return fmt.Sprintf("0198a000-0000-7000-8000-0000000000%02d", n)
}

func authzActors() []authzActor {
	wsA := authzWorkspace(1)
	return []authzActor{
		{"trainee(所属あり)", &domain.User{ID: 1, Role: domain.RoleTrainee, IsActive: true, WorkspaceID: &wsA}},
		{"trainee(未所属)", &domain.User{ID: 2, Role: domain.RoleTrainee, IsActive: true}},
		{"company_admin(所属あり)", &domain.User{ID: 3, Role: domain.RoleCompanyAdmin, IsActive: true, WorkspaceID: &wsA}},
		{"company_admin(未所属)", &domain.User{ID: 4, Role: domain.RoleCompanyAdmin, IsActive: true}},
		{"super_admin(未所属)", &domain.User{ID: 5, Role: domain.RoleSuperAdmin, IsActive: true}},
		{"未知のrole", &domain.User{ID: 6, Role: domain.RoleName("visitor"), IsActive: true}},
	}
}

// authzRequest は 1 本のリクエスト。body が空なら GET/DELETE 相当として送る。
type authzRequest struct {
	method string
	path   string
	body   string
}

func adminAuthzRequests() []authzRequest {
	return []authzRequest{
		{http.MethodGet, "/admin/members", ""},
		{http.MethodGet, "/admin/members/learning-summary", ""},
		{http.MethodPatch, "/admin/members/99999/active", `{"active":false}`},
		{http.MethodDelete, "/admin/members/99999", ""},
		{http.MethodGet, "/admin/invitations", ""},
		{http.MethodPost, "/admin/invitations", `{"email":"probe@example.test","role":"trainee"}`},
		{http.MethodDelete, "/admin/invitations/99999", ""},
		{http.MethodGet, "/admin/audit-events", ""},
	}
}

// newAdminAuthzRouter は本番の registerAdminRoutes をそのまま使い、認証だけを差し替える。
// JWT の検証と users の引き当ては本題ではないので、確定した actor を context に置く。
func newAdminAuthzRouter(t *testing.T, db *sql.DB, actor *domain.User) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	g := r.Group("")
	g.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyCurrentUserID, actor.ID)
		c.Set(middleware.ContextKeyCurrentUser, actor)
		c.Next()
	})

	deps := &routeDeps{db: db, cfg: &config.Config{}, userRepo: nil}
	registerAdminRoutes(g, deps, newAuditMiddleware(db))
	return r
}

// TestAdminEndpointAuthzMatrix_Integration は「誰がどの管理エンドポイントを通れるか」を
// 実測して固定する。表は現状の追認ではなく、**通ってはいけない組み合わせが 200 を返さない**
// ことの確認として読む。
func TestAdminEndpointAuthzMatrix_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)

	type cell struct {
		actor  string
		req    authzRequest
		status int
	}
	var cells []cell

	for _, a := range authzActors() {
		r := newAdminAuthzRouter(t, db, a.user)
		for _, req := range adminAuthzRequests() {
			var bodyReader *strings.Reader
			if req.body != "" {
				bodyReader = strings.NewReader(req.body)
			} else {
				bodyReader = strings.NewReader("")
			}
			httpReq := httptest.NewRequest(req.method, req.path, bodyReader)
			if req.body != "" {
				httpReq.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httpReq)
			cells = append(cells, cell{a.name, req, w.Code})
		}
	}

	// 実測表を残す。-v で読める形にして、変更のたびに差分を目で見られるようにする。
	sort.SliceStable(cells, func(i, j int) bool {
		if cells[i].req.path != cells[j].req.path {
			return cells[i].req.path < cells[j].req.path
		}
		return cells[i].actor < cells[j].actor
	})
	t.Log("=== 管理エンドポイントの認可マトリクス（実測）===")
	for _, c := range cells {
		t.Logf("%-28s %-7s %-40s -> %d", c.actor, c.req.method, c.req.path, c.status)
	}

	// 通ってはいけない組み合わせ: trainee と未知の role は、どの管理エンドポイントでも
	// 業務処理へ到達してはならない。到達したかどうかは「認可で弾かれたか」で見る。
	for _, c := range cells {
		blockedActor := strings.HasPrefix(c.actor, "trainee") || c.actor == "未知のrole"
		if !blockedActor {
			continue
		}
		if c.status != http.StatusForbidden && c.status != http.StatusUnauthorized {
			t.Errorf("%s が %s %s で弾かれていない: status=%d（403/401 のはず）",
				c.actor, c.req.method, c.req.path, c.status)
		}
	}
}

// insertAuthzUser は検証用のユーザーを 1 人作る。workspaceID が空なら未所属。
func insertAuthzUser(t *testing.T, db *sql.DB, email, workspaceID string, roleID int) uint64 {
	t.Helper()
	var ws any
	if workspaceID != "" {
		ws = workspaceID
	}
	var id uint64
	if err := db.QueryRow(
		`INSERT INTO users (email, name, workspace_id, role_id, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, true, now(), now()) RETURNING id`,
		email, email, ws, roleID,
	).Scan(&id); err != nil {
		t.Fatalf("ユーザー作成に失敗: %v", err)
	}
	return id
}

func insertAuthzWorkspace(t *testing.T, db *sql.DB, slug string) string {
	t.Helper()
	var id string
	if err := db.QueryRow(
		`INSERT INTO workspaces (id, slug, name, is_active) VALUES (gen_random_uuid(), $1, $1, true) RETURNING id`,
		slug,
	).Scan(&id); err != nil {
		t.Fatalf("ワークスペース作成に失敗: %v", err)
	}
	return id
}

// TestAdminMemberManagementCrossTenant_Integration は「実在する相手」に対する越境を実測する。
//
// マトリクスの検証は存在しない ID を叩いており、そこでは認可より先に存在確認が走って 404 に
// なる。つまり「弾かれた」ように見えても認可が効いた証拠にならない。実在する相手を用意して、
// 別ワークスペースの管理者が本当に手を出せないことを確かめる。
func TestAdminMemberManagementCrossTenant_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	testsupport.TruncateAll(t, db, "users", "workspaces")

	wsA := insertAuthzWorkspace(t, db, "authz-a")
	wsB := insertAuthzWorkspace(t, db, "authz-b")

	const roleCompanyAdmin, roleTrainee = 2, 3
	adminA := insertAuthzUser(t, db, "admin-a@example.test", wsA, roleCompanyAdmin)
	victimB := insertAuthzUser(t, db, "victim-b@example.test", wsB, roleTrainee)

	actor := &domain.User{
		ID: adminA, Role: domain.RoleCompanyAdmin, IsActive: true, WorkspaceID: &wsA,
	}
	r := newAdminAuthzRouter(t, db, actor)

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"停止", http.MethodPatch, fmt.Sprintf("/admin/members/%d/active", victimB), `{"active":false}`},
		{"削除", http.MethodDelete, fmt.Sprintf("/admin/members/%d", victimB), ""},
	}
	for _, c := range cases {
		t.Run("別ワークスペースのメンバーを"+c.name+"できない", func(t *testing.T) {
			req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
			if c.body != "" {
				req.Header.Set("Content-Type", "application/json")
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code == http.StatusOK || w.Code == http.StatusNoContent {
				t.Fatalf("越境できてしまった: status=%d body=%s", w.Code, w.Body.String())
			}
			t.Logf("%s -> %d %s", c.name, w.Code, strings.TrimSpace(w.Body.String()))

			// 実際に DB が書き換わっていないことまで見る（応答だけ見ると、
			// エラーを返しつつ更新している実装を見逃す）。
			var active bool
			var deletedAt sql.NullTime
			if err := db.QueryRow(
				`SELECT is_active, deleted_at FROM users WHERE id = $1`, victimB,
			).Scan(&active, &deletedAt); err != nil {
				t.Fatalf("対象の状態を読めない: %v", err)
			}
			if !active || deletedAt.Valid {
				t.Fatalf("弾いたはずなのに対象が変更されている: is_active=%v deleted_at=%v", active, deletedAt)
			}
		})
	}

	// 一覧が他ワークスペースの人を含まないことも見る。
	t.Run("一覧に別ワークスペースのメンバーは出ない", func(t *testing.T) {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/members", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "victim-b@example.test") {
			t.Fatalf("別ワークスペースのメンバーが一覧に出ている: %s", w.Body.String())
		}
	})
}
