package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AuthHandler は Cognito 関連の認証エンドポイントを提供する。
// Spring Boot の CognitoAuthController に相当。
type AuthHandler struct {
	getCurrentUser *usecase.GetCurrentUserUseCase
}

func NewAuthHandler(getCurrentUser *usecase.GetCurrentUserUseCase) *AuthHandler {
	return &AuthHandler{getCurrentUser: getCurrentUser}
}

// Me は現在ログイン中のユーザー情報を返す。
func (h *AuthHandler) Me(c *gin.Context) {
	sub, ok := c.Get(middleware.ContextKeyCognitoSub)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, err := h.getCurrentUser.Execute(c.Request.Context(), sub.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user_not_found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Logout はリフレッシュ・アクセストークンの Cookie を消去する。
func (h *AuthHandler) Logout(c *gin.Context) {
	c.SetCookie(middleware.CookieAccessToken, "", -1, "/", "", true, true)
	c.SetCookie("refresh_token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "ログアウトしました。"})
}
