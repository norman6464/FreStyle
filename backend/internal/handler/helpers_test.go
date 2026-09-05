package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

func TestActorWorkspaceFromContext(t *testing.T) {
	t.Run("認証済み user から id/所属ワークスペースを取り出す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		workspaceID := "ws-7"
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 42, WorkspaceID: &workspaceID})

		uid, workspace, ok := actorWorkspaceFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, uint64(42), uid)
		gotID, affiliated := workspace.WorkspaceID()
		assert.True(t, affiliated)
		assert.Equal(t, "ws-7", gotID)
	})

	t.Run("ワークスペース未所属(nil)なら未所属の WorkspaceRef", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, WorkspaceID: nil})

		_, workspace, ok := actorWorkspaceFromContext(c)

		assert.True(t, ok)
		_, affiliated := workspace.WorkspaceID()
		assert.False(t, affiliated)
	})

	t.Run("未認証なら 401 を書き ok=false", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, _, ok := actorWorkspaceFromContext(c)

		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserWorkspaceRef(t *testing.T) {
	wid := "ws-9"
	assert.Equal(t, domain.WorkspaceRefOf("ws-9"), domain.User{WorkspaceID: &wid}.WorkspaceRef())
	assert.Equal(t, domain.NoWorkspace(), domain.User{WorkspaceID: nil}.WorkspaceRef())
}
