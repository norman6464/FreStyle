package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/bedrock"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// routeDeps はドメインごとの register*Routes 関数に渡す共通依存。
type routeDeps struct {
	db            *gorm.DB
	cfg           *config.Config
	userRepo      repository.UserRepository
	bedrockClient *bedrock.Client
	msgRepo       repository.AiChatMessageRepository
}

// NewRouter は API ルーティングを組み立てる。
func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend"})
	})

	ctx := context.Background()

	bc, err := bedrock.NewClient(ctx, cfg.Bedrock.Region, cfg.Bedrock.ModelID)
	if err != nil {
		log.Printf("WARN: Bedrock client init failed (AI chat WS will be unavailable): %v", err)
	}

	msgRepo, err := repository.NewAiChatMessageRepository(ctx, cfg.DynamoDB.Region, cfg.DynamoDB.AiChatTable)
	if err != nil {
		log.Printf("WARN: DynamoDB client init failed (AI chat WS will be unavailable): %v", err)
	}

	deps := &routeDeps{
		db:            db,
		cfg:           cfg,
		userRepo:      repository.NewUserRepository(db),
		bedrockClient: bc,
		msgRepo:       msgRepo,
	}

	v2 := r.Group("/api/v2")

	registerHealthRoutes(v2, deps)
	authHandler := registerAuthPublicRoutes(v2, deps)

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
	registerEmbedRoutes(authed)
	registerPhpRoutes(authed, deps)
	registerWebSocketRoutes(authed, deps)

	return r
}

// registerHealthRoutes は認証不要のヘルスチェック (/api/v2/health) を登録する。
func registerHealthRoutes(g *gin.RouterGroup, deps *routeDeps) {
	h := NewHealthHandler(usecase.NewCheckHealthUseCase(repository.NewHealthRepository(deps.db)))
	g.GET("/health", h.Get)
}
