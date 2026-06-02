package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/infra/embed"
)

// registerEmbedRoutes は外部 URL の OGP / oEmbed メタ取得エンドポイントを登録する。
func registerEmbedRoutes(g *gin.RouterGroup) {
	h := NewEmbedHandler(embed.NewFetcher())
	g.GET("/embeds/oembed", h.Resolve)
}
