package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

// runActorWorkspace は本番と同じく gin のルーティングを通して actorWorkspaceFromContext を
// 実行する。gin.CreateTestContext はルーティング経路を通らないため、未認証で 401 を書いた
// あとに後続の処理が止まること（Abort が効くこと）まで確かめられない。
//
// setUser が nil のときは current user を注入せず、未認証を再現する。
func runActorWorkspace(setUser func(c *gin.Context)) (rec *httptest.ResponseRecorder, got actorWorkspaceResult) {
	r := gin.New()
	r.GET("/probe", func(c *gin.Context) {
		if setUser != nil {
			setUser(c)
		}
		uid, workspace, role, ok := actorWorkspaceFromContext(c)
		got = actorWorkspaceResult{userID: uid, workspace: workspace, role: role, ok: ok}
		if !ok {
			return // 本番の handler と同じく早期 return する
		}
		got.reachedEnd = true
		c.Status(http.StatusOK)
	})
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/probe", nil))
	return rec, got
}

// actorWorkspaceResult は runActorWorkspace が持ち帰る観測値。
type actorWorkspaceResult struct {
	userID    uint64
	workspace domain.WorkspaceRef
	role      domain.RoleName
	ok        bool
	// reachedEnd は ok=true のときだけ handler の末尾まで到達したことを表す。
	reachedEnd bool
}

func TestActorWorkspaceFromContext(t *testing.T) {
	t.Run("認証済み user から id/所属ワークスペース/role を取り出す", func(t *testing.T) {
		workspaceID := "ws-7"
		rec, got := runActorWorkspace(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 42, WorkspaceID: &workspaceID, Role: domain.RoleCompanyAdmin})
		})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, got.ok)
		assert.True(t, got.reachedEnd)
		assert.Equal(t, uint64(42), got.userID)
		gotID, affiliated := got.workspace.WorkspaceID()
		assert.True(t, affiliated)
		assert.Equal(t, "ws-7", gotID)
		assert.Equal(t, domain.RoleCompanyAdmin, got.role)
	})

	t.Run("ワークスペース未所属(nil)なら未所属の WorkspaceRef", func(t *testing.T) {
		rec, got := runActorWorkspace(func(c *gin.Context) {
			c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, WorkspaceID: nil, Role: domain.RoleSuperAdmin})
		})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, got.ok)
		_, affiliated := got.workspace.WorkspaceID()
		assert.False(t, affiliated)
	})

	t.Run("未認証なら 401 を書き ok=false", func(t *testing.T) {
		rec, got := runActorWorkspace(nil)

		assert.False(t, got.ok)
		assert.False(t, got.reachedEnd, "401 のあとに handler の続きが動いている")
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestRespondEntityErr(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"レコード未検出は 404", domain.ErrNotFound, http.StatusNotFound},
		{"forbidden は 403", errors.New("forbidden"), http.StatusForbidden},
		{"forbidden 詳細付きも 403", errors.New("forbidden: only company_admin or super_admin can create materials"), http.StatusForbidden},
		{"テナント未所属は 403", errors.New("actor must belong to a workspace"), http.StatusForbidden},
		{"その他は 500", errors.New("db down"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			respondEntityErr(c, tc.err, "見つかりません", "失敗しました")

			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestUserWorkspaceRef(t *testing.T) {
	wid := "ws-9"
	assert.Equal(t, domain.WorkspaceRefOf("ws-9"), domain.User{WorkspaceID: &wid}.WorkspaceRef())
	assert.Equal(t, domain.NoWorkspace(), domain.User{WorkspaceID: nil}.WorkspaceRef())
}
