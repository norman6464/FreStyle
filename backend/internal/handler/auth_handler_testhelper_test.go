package handler

import (
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

// mustNewRequest はテスト用に最小の *http.Request を返す。
// auth_handler 配下の handler は c.Request.Context() しか参照しないため、URL / method は何でも良い。
func mustNewRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/", nil)
}

// upsertAllowed は upsertUserFromIDToken が user を解決できたかだけを取り出すテスト補助。
// 内部エラー(err!=nil)の切り分けは password_login_test.go の handler テストで検証する。
func upsertAllowed(h *AuthHandler, c *gin.Context, idToken string) bool {
	user, _ := h.upsertUserFromIDToken(c, idToken, "")
	return user != nil
}
