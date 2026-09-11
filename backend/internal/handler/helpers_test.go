package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

// testCtx は handler の単体テスト（httptest ベース）が共有する gin.Context の組み立て。
// rich_text_image_handler_test.go / code_execute_handler_test.go が参照し続けるため、
// 共通の test helper 置き場であるこのファイルへ移設した。
func testCtx(method, body string, uid uint64, idVal string) (*httptest.ResponseRecorder, *gin.Context) {
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
