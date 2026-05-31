package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の Cognito 認証エンドポイント（logout / callback / refresh-token）を
// 登録し、authed group で再利用するため AuthHandler を返す。callback / refresh は Cookie を発行・更新するため JWTAuth 対象外。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	invitations := persistence.NewAdminInvitationRepository(deps.db)
	authHandler := NewAuthHandler(getCurrentUser, deps.userRepo, invitations, &deps.cfg.Cognito)

	g.POST("/auth/cognito/logout", authHandler.Logout)
	g.POST("/auth/cognito/callback", authHandler.Callback)
	g.POST("/auth/cognito/refresh-token", authHandler.Refresh)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
