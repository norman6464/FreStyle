package handler

import (
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
)

// ctxJSON は JSON ボディ・パス params・current user を仕込んだテスト用の gin.Context を返す。
// 複数の handler テストで共有する汎用ヘルパー。
func ctxJSON(method, body string, params gin.Params, user *domain.User) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = params
	if user != nil {
		c.Set(middleware.ContextKeyCurrentUser, user)
	}
	return w, c
}

// idParam は gin の :id パスパラメータを模す。
func idParam(v string) gin.Params { return gin.Params{{Key: "id", Value: v}} }
