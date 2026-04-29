package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の Cognito 認証エンドポイント (logout / callback / refresh-token)
// を登録し、後段の authed group で再利用するため AuthHandler を返す。
//
// callback / refresh-token は HttpOnly Cookie の発行・更新を行うため middleware.JWTAuth の対象外。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	authHandler := NewAuthHandler(getCurrentUser, deps.userRepo, &deps.cfg.Cognito)

	g.POST("/auth/cognito/logout", authHandler.Logout)
	// callback は code を受け取って token に交換するので認証不要
	g.POST("/auth/cognito/callback", authHandler.Callback)
	// refresh-token は HttpOnly Cookie の refresh_token を読むため認証 middleware の対象外
	g.POST("/auth/cognito/refresh-token", authHandler.Refresh)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
