package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の認証エンドポイント（login / logout / refresh）を
// 登録し、authed group で再利用するため AuthHandler を返す。login / refresh は Cookie を発行・更新するため JWTAuth 対象外。
// パスは provider 非依存の REST 形（/auth/login = 認可コード→token 交換 / /auth/logout / /auth/refresh）で、frontend の apiRoutes と一致させる。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	invitations := persistence.NewAdminInvitationRepository(deps.db)
	authHandler := NewAuthHandler(getCurrentUser, deps.userRepo, invitations, &deps.cfg.Cognito)

	g.POST("/auth/logout", authHandler.Logout)
	// login（認可コード→token 交換）は認証不要のため、総当たり緩和に per-IP 制限を掛ける。
	g.POST("/auth/login", middleware.RateLimitPerMinute(30, 10), authHandler.Callback)
	// refresh は正規ユーザーが定期的に叩くため、NAT 共有 IP を考慮して緩めに設定する。
	g.POST("/auth/refresh", middleware.RateLimitPerMinute(60, 30), authHandler.Refresh)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
