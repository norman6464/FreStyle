package handler

import (
	"github.com/gin-gonic/gin"
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
	{
		healthHandler := NewHealthHandler(
			usecase.NewCheckHealthUseCase(repository.NewHealthRepository(db)),
		)
		v2.GET("/health", healthHandler.Get)
	}

	return r
}
