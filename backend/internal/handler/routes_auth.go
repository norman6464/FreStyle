package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// registerAuthPublicRoutes は認証不要の login を登録し、
// authed group で再利用するため AuthHandler を返す。
// login は Bearer の ID トークン自体を検証するため JWTAuth の対象外
// （JWTAuth を先に通すと、まだ users 行の無い初回サインインが弾かれてしまう）。
//
// メールとパスワードをアプリのフォームで受ける経路（旧発行者の専用ログイン・
// パスワード再設定 API）は撤去した。パスワードを受け取るのは発行者の
// サインイン画面（Firebase JS SDK）の役目で、アプリが受け取ると、二要素・ロックアウト・
// パスワードの強さといった発行者側の守りをすべて素通りする経路を自分で開くことになる。
//
// logout / refresh は撤去した。GCIP はサーバー側の Cookie セッションを持たず、
// サインアウトもトークンの自動更新もクライアント SDK 側で完結する。
func registerAuthPublicRoutes(g *gin.RouterGroup, deps *routeDeps) *AuthHandler {
	getCurrentUser := user.NewGetCurrentUserUseCase(deps.userRepo)
	upsertUser := user.NewUpsertUserFromIDTokenUseCase(
		deps.userRepo,
		persistence.NewUserOidcIdentityRepository(deps.db),
		persistence.NewTxManager(deps.db),
	)
	ensurePersonalWorkspace := kb.NewEnsurePersonalWorkspaceUseCase(
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewWorkspaceProvisioner(deps.db),
	)

	authHandler := NewAuthHandler(
		getCurrentUser, upsertUser, ensurePersonalWorkspace, deps.verifier,
	)

	// login（ID トークンの検証+upsert）は認証不要のため、総当たり緩和に per-IP 制限を掛ける。
	g.POST("/auth/login", middleware.RateLimitPerMinute(30, 10), authHandler.Login)

	return authHandler
}

// registerAuthAuthedRoutes は認証必須の自己情報取得 (/auth/me) を登録する。
func registerAuthAuthedRoutes(g *gin.RouterGroup, authHandler *AuthHandler) {
	g.GET("/auth/me", authHandler.Me)
}
