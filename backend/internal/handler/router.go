package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// routeDeps はドメインごとの register*Routes 関数に渡す共通依存。
// db / cfg / userRepo の 3 点が大半の register で必要になるため struct でまとめる。
// 個別の register が追加で repository を必要とする場合は db から都度組み立てる
// (lifecycle はリクエスト境界より長く、Gin ハンドラからは pure な関数 closure として参照される)。
type routeDeps struct {
	db       *gorm.DB
	cfg      *config.Config
	userRepo repository.UserRepository
}

// NewRouter は API ルーティングを組み立てる。
// /api/v2/* は Go バックエンドの単独エンドポイント（旧 Spring Boot /api/* は廃止済み）。
//
// 旧実装は本関数 1 つに 28 phase 分のルーティング登録が直書きされていたが (315 行)、
// 1 ファイル 1 ドメインの routes_*.go に分割し、本関数は組み立てだけを担う。
func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	// 全エンドポイントで CORS を許可（normanblog.com など allowlisted origin）
	r.Use(middleware.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend"})
	})

	deps := &routeDeps{
		db:       db,
		cfg:      cfg,
		userRepo: repository.NewUserRepository(db),
	}

	v2 := r.Group("/api/v2")

	// 認証不要のルート（health / cognito callback / refresh）。
	registerHealthRoutes(v2, deps)
	authHandler := registerAuthPublicRoutes(v2, deps)

	// 認証必須ルート。JWTAuth で sub を context に詰めた後、CurrentUser で users.id を解決する。
	authed := v2.Group("")
	authed.Use(middleware.JWTAuth())
	authed.Use(middleware.CurrentUser(deps.userRepo))

	registerAuthAuthedRoutes(authed, authHandler)
	registerChatRoutes(authed, deps)
	registerProfileRoutes(authed, deps)
	registerPracticeRoutes(authed, deps)
	registerNoteRoutes(authed, deps)
	registerScoreRoutes(authed, deps)
	registerPhraseRoutes(authed, deps)
	registerSocialRoutes(authed, deps)
	registerSettingsRoutes(authed, deps)
	registerAdminRoutes(authed, deps)
	registerWebSocketRoutes(authed)

	return r
}

// registerHealthRoutes は認証不要のヘルスチェック (/api/v2/health) を登録する。
func registerHealthRoutes(g *gin.RouterGroup, deps *routeDeps) {
	h := NewHealthHandler(usecase.NewCheckHealthUseCase(repository.NewHealthRepository(deps.db)))
	g.GET("/health", h.Get)
}
