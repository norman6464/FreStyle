package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// NewRouter は API ルーティングを組み立てる。
// /api/v2/* は Spring Boot の /api/* と並行運用する Go 側のエンドポイント。
func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend"})
	})

	v2 := r.Group("/api/v2")

	// Phase 1: 認証不要のヘルスチェック
	healthHandler := NewHealthHandler(
		usecase.NewCheckHealthUseCase(repository.NewHealthRepository(db)),
	)
	v2.GET("/health", healthHandler.Get)

	// Phase 2: 認証 (Cognito)
	userRepo := repository.NewUserRepository(db)
	authHandler := NewAuthHandler(usecase.NewGetCurrentUserUseCase(userRepo))
	v2.POST("/auth/cognito/logout", authHandler.Logout)

	// 認証必須グループ
	authed := v2.Group("")
	authed.Use(middleware.JWTAuth())
	authed.GET("/auth/me", authHandler.Me)

	// Phase 3: AI チャット
	aiSessionRepo := repository.NewAiChatSessionRepository(db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
	)
	authed.GET("/ai-chat/sessions", aiHandler.GetSessions)
	authed.POST("/ai-chat/sessions", aiHandler.CreateSession)

	return r
}
