package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

// noteCtx は handler の単体テスト（httptest ベース）が共有する gin.Context の組み立て。
// 元々は note_handler_test.go にあったが、その削除後も note_image_handler_test.go /
// code_execute_handler_test.go が参照し続けるため、共通の test helper 置き場である
// このファイルへ移設した。
func noteCtx(method, body string, uid uint64, idVal string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if uid != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, uid)
	}
	if idVal != "" {
		c.Params = gin.Params{{Key: "id", Value: idVal}}
	}
	return w, c
}

func TestUserWorkspaceRef(t *testing.T) {
	wid := "ws-9"
	assert.Equal(t, domain.WorkspaceRefOf("ws-9"), domain.User{WorkspaceID: &wid}.WorkspaceRef())
	assert.Equal(t, domain.NoWorkspace(), domain.User{WorkspaceID: nil}.WorkspaceRef())
}
