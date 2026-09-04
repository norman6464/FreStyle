//go:build integration

package handler

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 教材の権限操作 API を、実 PostgreSQL・本番と同じ配線で確かめる。
//
// 認可は usecase が持つので、効いているかどうかは HTTP の入口を実際に叩かないと
// 分からない。ここで固定するのは:
//
//  1. 管理できない相手は 1 本も通らないこと
//  2. **見えない相手には 404、見えているが管理できない相手には 403** であること
//     （前者を撃ち分けると ID の総当たりで教材の実在が分かる）
//  3. 管理できる相手なら通り、通した結果が実効権限に反映されること

type materialGrantEnv struct {
	db      *sql.DB
	ws      string
	course  uint64
	chapter uint64
	// owner はコースの admin。target へ権限を張る側。
	owner uint64
	// reader は一員だが付与を持たない（公開済みは読めるが管理はできない）。
	reader uint64
	// outsider はワークスペースに所属していない。
	outsider        uint64
	targetPrincipal string
}

func newMaterialGrantEnv(t *testing.T, db *sql.DB) *materialGrantEnv {
	t.Helper()
	testsupport.TruncateAll(t, db, append([]string{"course_chapters", "courses"}, kbIntegrationTables...)...)

	e := &materialGrantEnv{db: db, course: 8001, chapter: 8101}
	e.ws = kbInsertWorkspace(t, db, "material-grant")
	e.owner = kbInsertUser(t, db, "owner")
	e.reader = kbInsertUser(t, db, "reader")
	e.outsider = kbInsertUser(t, db, "outsider")
	target := kbInsertUser(t, db, "target")

	_, err := db.Exec(
		`INSERT INTO courses (id, workspace_id, created_by_user_id, title, is_published, created_at, updated_at)
		 VALUES ($1, $2, $3, 'コース', true, now(), now())`, e.course, e.ws, e.owner,
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title, is_published, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, '章', true, now(), now())`, e.chapter, e.ws, e.course, e.owner,
	)
	require.NoError(t, err)

	// 所属は principals の行が唯一の表現。outsider だけ作らない。
	perms := persistence.NewKnowledgeBasePermissionRepository(db)
	ownerPrincipal, err := perms.EnsureUserPrincipal(t.Context(), e.ws, e.owner)
	require.NoError(t, err)
	_, err = perms.EnsureUserPrincipal(t.Context(), e.ws, e.reader)
	require.NoError(t, err)
	targetPrincipal, err := perms.EnsureUserPrincipal(t.Context(), e.ws, target)
	require.NoError(t, err)
	e.targetPrincipal = targetPrincipal.ID

	// owner だけがコースの admin。
	_, err = persistence.NewMaterialPermissionRepository(db).
		UpsertCourseGrant(t.Context(), e.ws, e.course, ownerPrincipal.ID, domain.GrantRoleAdmin)
	require.NoError(t, err)
	return e
}

// as は指定ユーザーとしてリクエストを流すルータを返す（本番と同じ登録経路を通す）。
func (e *materialGrantEnv) as(t *testing.T, userID uint64, affiliated bool) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/api/v2")
	g.Use(func(c *gin.Context) {
		user := &domain.User{ID: userID}
		if affiliated {
			ws := e.ws
			user.WorkspaceID = &ws
		}
		c.Set(middleware.ContextKeyCurrentUserID, userID)
		c.Set(middleware.ContextKeyCurrentUser, user)
		c.Next()
	})
	deps := &routeDeps{db: e.db}
	registerCourseRoutes(g, deps)
	registerTeachingMaterialRoutes(g, deps)
	return r
}

func (e *materialGrantEnv) do(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// materialGrantRoute は権限を変える 1 経路。
type materialGrantRoute struct {
	name     string
	method   string
	path     func(e *materialGrantEnv) string
	body     string
	okStatus int
}

func materialGrantRoutes() []materialGrantRoute {
	return []materialGrantRoute{
		{
			name: "コースの権限一覧", method: http.MethodGet, okStatus: http.StatusOK,
			path: func(e *materialGrantEnv) string { return courseGrantsPath(e.course) },
		},
		{
			name: "コースの権限付与", method: http.MethodPut, okStatus: http.StatusOK,
			path: func(e *materialGrantEnv) string { return courseGrantsPath(e.course) + "/" + e.targetPrincipal },
			body: `{"role":"editor"}`,
		},
		{
			name: "コースの権限取り消し", method: http.MethodDelete, okStatus: http.StatusNoContent,
			path: func(e *materialGrantEnv) string { return courseGrantsPath(e.course) + "/" + e.targetPrincipal },
		},
		{
			name: "相手の一覧", method: http.MethodGet, okStatus: http.StatusOK,
			path: func(e *materialGrantEnv) string { return coursePath(e.course) + "/principals" },
		},
		{
			name: "教材の権限一覧", method: http.MethodGet, okStatus: http.StatusOK,
			path: func(e *materialGrantEnv) string { return chapterGrantsPath(e.chapter) },
		},
		{
			name: "教材の権限付与", method: http.MethodPut, okStatus: http.StatusOK,
			path: func(e *materialGrantEnv) string { return chapterGrantsPath(e.chapter) + "/" + e.targetPrincipal },
			body: `{"role":"editor"}`,
		},
		{
			name: "教材の権限取り消し", method: http.MethodDelete, okStatus: http.StatusNoContent,
			path: func(e *materialGrantEnv) string { return chapterGrantsPath(e.chapter) + "/" + e.targetPrincipal },
		},
	}
}

func coursePath(id uint64) string       { return "/api/v2/courses/" + strconv.FormatUint(id, 10) }
func courseGrantsPath(id uint64) string { return coursePath(id) + "/grants" }
func chapterGrantsPath(id uint64) string {
	return "/api/v2/teaching-materials/" + strconv.FormatUint(id, 10) + "/grants"
}

func TestMaterialGrantAPI_管理できない相手は1本も通らない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)

	t.Run("読めるが管理できない相手は403", func(t *testing.T) {
		// 実在は既に知っているので理由を返してよい。
		for _, route := range materialGrantRoutes() {
			t.Run(route.name, func(t *testing.T) {
				e := newMaterialGrantEnv(t, sqlDB)
				w := e.do(t, e.as(t, e.reader, true), route.method, route.path(e), route.body)
				assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
			})
		}
	})

	t.Run("所属していない相手は404（実在を教えない）", func(t *testing.T) {
		for _, route := range materialGrantRoutes() {
			t.Run(route.name, func(t *testing.T) {
				e := newMaterialGrantEnv(t, sqlDB)
				w := e.do(t, e.as(t, e.outsider, false), route.method, route.path(e), route.body)
				assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
			})
		}
	})
}

func TestMaterialGrantAPI_管理できる相手は全経路を通れる_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	for _, route := range materialGrantRoutes() {
		t.Run(route.name, func(t *testing.T) {
			e := newMaterialGrantEnv(t, sqlDB)
			w := e.do(t, e.as(t, e.owner, true), route.method, route.path(e), route.body)
			assert.Equal(t, route.okStatus, w.Code, w.Body.String())
		})
	}
}

func TestMaterialGrantAPI_張った権限がそのまま実効権限になる_Integration(t *testing.T) {
	// 配線だけして書けていない、を防ぐ。付与のあと、その人が実際に編集できること。
	sqlDB := testsupport.OpenTestDB(t)
	e := newMaterialGrantEnv(t, sqlDB)
	perms := persistence.NewMaterialPermissionRepository(sqlDB)

	target := kbInsertUser(t, sqlDB, "grantee")
	p, err := persistence.NewKnowledgeBasePermissionRepository(sqlDB).
		EnsureUserPrincipal(t.Context(), e.ws, target)
	require.NoError(t, err)

	canEdit := func() bool {
		facts, err := perms.ChapterFactsForUser(t.Context(), e.ws, e.chapter, target)
		require.NoError(t, err)
		return domain.ResolveMaterialPermission(*facts).CanEdit
	}
	require.False(t, canEdit(), "前提: まだ編集できない")

	owner := e.as(t, e.owner, true)
	w := e.do(t, owner, http.MethodPut, courseGrantsPath(e.course)+"/"+p.ID, `{"role":"editor"}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.True(t, canEdit(), "コースへ張った付与が章まで届いていない")

	w = e.do(t, owner, http.MethodDelete, courseGrantsPath(e.course)+"/"+p.ID, "")
	require.Equal(t, http.StatusNoContent, w.Code)
	assert.False(t, canEdit(), "取り消したのに編集できたまま")
}

func TestMaterialGrantAPI_既知でない役割は400_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	e := newMaterialGrantEnv(t, sqlDB)
	w := e.do(t, e.as(t, e.owner, true), http.MethodPut,
		courseGrantsPath(e.course)+"/"+e.targetPrincipal, `{"role":"owner"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}

func TestMaterialGrantAPI_別テナントの主体には張れない_Integration(t *testing.T) {
	sqlDB := testsupport.OpenTestDB(t)
	e := newMaterialGrantEnv(t, sqlDB)
	otherWS := kbInsertWorkspace(t, sqlDB, "material-grant-other")
	stranger := kbInsertUser(t, sqlDB, "stranger")
	p, err := persistence.NewKnowledgeBasePermissionRepository(sqlDB).
		EnsureUserPrincipal(t.Context(), otherWS, stranger)
	require.NoError(t, err)

	// FK 違反（500）ではなく「対象が無い」として返ること。
	w := e.do(t, e.as(t, e.owner, true), http.MethodPut,
		courseGrantsPath(e.course)+"/"+p.ID, `{"role":"editor"}`)
	assert.Equal(t, http.StatusNotFound, w.Code, w.Body.String())
}
