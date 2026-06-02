package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の Cognito 認証エンドポイント (logout / callback / refresh-token)
// を登録し、後段の authed group で再利用するため AuthHandler を返す。
//
// callback / refresh-token は HttpOnly Cookie の発行・更新を行うため middleware.JWTAuth の対象外。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	invitations := persistence.NewAdminInvitationRepository(deps.db)
	authHandler := NewAuthHandler(getCurrentUser, deps.userRepo, invitations, &deps.cfg.Cognito)

	g.POST("/auth/cognito/logout", authHandler.Logout)
	// callback は code を受け取って token に交換するので認証不要。総当たり緩和に per-IP 制限を掛ける。
	g.POST("/auth/cognito/callback", middleware.RateLimitPerMinute(30, 10), authHandler.Callback)
	// refresh-token は正規ユーザーが定期的に叩くため、NAT 共有 IP を考慮して緩めに設定する。
	g.POST("/auth/cognito/refresh-token", middleware.RateLimitPerMinute(60, 30), authHandler.Refresh)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
