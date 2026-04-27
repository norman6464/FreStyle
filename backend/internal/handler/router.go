package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter は API ルーティングを組み立てる。
// Phase 0 では空のルータのみを返し、Phase 1 以降でハンドラを順次追加する。
func NewRouter(_ *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend (Phase 0 bootstrap)"})
	})

	return r
}
