package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の認証エンドポイント（login / cognito login / logout / refresh）を
// 登録し、authed group で再利用するため AuthHandler を返す。login / refresh は Cookie を発行・更新するため JWTAuth 対象外。
// パスは Hosted UI 認可コード交換が /auth/login、アプリ内メール/パスワードが /auth/cognito/login で frontend の apiRoutes と一致させる。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	invitations := persistence.NewAdminInvitationRepository(deps.db)
	aiAccess := usecase.NewAiChatEnabledForUserUseCase(persistence.NewCompanyRepository(deps.db))

	// USER_PASSWORD_AUTH 用の authenticator。AWS 認証情報の解決に失敗しても起動は止めず、
	// nil のまま渡して /auth/cognito/login だけ 500 にする（Hosted UI ログインには影響させない）。
	var pwAuth passwordAuthenticator
	if pa, err := cognito.NewPasswordAuthenticator(context.Background(), deps.cfg.Cognito.Region, deps.cfg.Cognito.ClientID, deps.cfg.Cognito.ClientSecret); err != nil {
		log.Printf("password authenticator init failed: %v", err)
	} else {
		pwAuth = pa
	}

	authHandler := NewAuthHandler(getCurrentUser, deps.userRepo, invitations, &deps.cfg.Cognito, pwAuth, aiAccess)

	g.POST("/auth/logout", authHandler.Logout)
	// login（認可コード→token 交換）は認証不要のため、総当たり緩和に per-IP 制限を掛ける。
	g.POST("/auth/login", middleware.RateLimitPerMinute(30, 10), authHandler.Callback)
	// cognito/login（メール/パスワード）はパスワード総当たり面なので callback より厳しめに絞る。
	g.POST("/auth/cognito/login", middleware.RateLimitPerMinute(10, 5), authHandler.Login)
	// refresh は正規ユーザーが定期的に叩くため、NAT 共有 IP を考慮して緩めに設定する。
	g.POST("/auth/refresh", middleware.RateLimitPerMinute(60, 30), authHandler.Refresh)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
