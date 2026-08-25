package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/infra/localauth"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAuthPublicRoutes は認証不要の認証エンドポイント（login / cognito login / logout / refresh）を
// 登録し、authed group で再利用するため AuthHandler を返す。login / refresh は Cookie を発行・更新するため JWTAuth 対象外。
// パスは Hosted UI 認可コード交換が /auth/login、アプリ内メール/パスワードが /auth/cognito/login で frontend の apiRoutes と一致させる。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := usecase.NewGetCurrentUserUseCase(deps.userRepo)
	invitations := persistence.NewAdminInvitationRepository(deps.db)
	aiAccess := usecase.NewAiChatEnabledForUserUseCase(persistence.NewCompanyRepository(deps.db))
	upsertUser := usecase.NewUpsertUserFromIDTokenUseCase(
		deps.userRepo,
		invitations,
		deps.cfg.BootstrapSuperAdminEmail,
	)
	promoteAdmin := usecase.NewPromoteCognitoAdminRoleUseCase(deps.userRepo)
	platformAdmin := usecase.NewSyncPlatformAdminUseCase(deps.userRepo)

	pwAuth := buildPasswordAuthenticator(deps)

	authHandler := NewAuthHandler(getCurrentUser, upsertUser, promoteAdmin, platformAdmin, &deps.cfg.Cognito, pwAuth, aiAccess)

	g.POST("/auth/logout", authHandler.Logout)
	// login（認可コード→token 交換）は認証不要のため、総当たり緩和に per-IP 制限を掛ける。
	g.POST("/auth/login", middleware.RateLimitPerMinute(30, 10), authHandler.Callback)
	// cognito/login（メール/パスワード）はパスワード総当たり面なので callback より厳しめに絞る。
	g.POST("/auth/cognito/login", middleware.RateLimitPerMinute(10, 5), authHandler.Login)
	// 新パスワード設定（一時パスワードでの初回ログイン）。login と同じくパスワード面なので絞る。
	g.POST("/auth/cognito/new-password", middleware.RateLimitPerMinute(10, 5), authHandler.NewPassword)
	// refresh は正規ユーザーが定期的に叩くため、NAT 共有 IP を考慮して緩めに設定する。
	g.POST("/auth/refresh", middleware.RateLimitPerMinute(60, 30), authHandler.Refresh)

	return authHandler
}

// buildPasswordAuthenticator は /auth/cognito/login のパスワード検証実装を選ぶ。
//   - LOCAL_PASSWORD_AUTH 有効 + APP_ENV=local: DB の bcrypt ハッシュで検証（infra/localauth）。
//     Cognito 不要でログインでき、seed ユーザーでもログインできる（FRESTYLE-311）。
//   - LOCAL_PASSWORD_AUTH 有効 + 非 local: fail closed（ERROR ログを出して Cognito 経路に固定）。
//   - それ以外: 従来どおり Cognito の USER_PASSWORD_AUTH。
//
// AWS 認証情報の解決に失敗しても起動は止めず、nil のまま渡して /auth/cognito/login だけ
// 500 にする（Hosted UI ログインには影響させない）。
func buildPasswordAuthenticator(deps *routeDeps) passwordAuthenticator {
	// cfg.LocalPasswordAuth は「LOCAL_PASSWORD_AUTH + 明示 APP_ENV=local + JWKS 未設定」を
	// すべて満たしたときだけ true（config.Load で解決済み）。localauth.New の appEnv ガードは
	// 二重の安全弁として残す。
	if deps.cfg.LocalPasswordAuth {
		if la, err := localauth.New(deps.userRepo, deps.cfg.AppEnv); err != nil {
			log.Printf("ERROR: localauth init failed: %v", err)
		} else {
			log.Printf("WARN: ローカル専用のパスワードログイン（DB bcrypt 検証）を使用します（本番では設定禁止）")
			return la
		}
	}
	pa, err := cognito.NewPasswordAuthenticator(context.Background(), deps.cfg.Cognito.Region, deps.cfg.Cognito.ClientID, deps.cfg.Cognito.ClientSecret)
	if err != nil {
		log.Printf("password authenticator init failed: %v", err)
		return nil
	}
	return pa
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
