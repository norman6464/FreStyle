package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/infra/embed"
)

// registerEmbedRoutes は外部 URL の OGP / oEmbed メタ情報取得エンドポイントを登録する。
// Note エディタの埋め込みカード機能（Markdown 互換拡張 PR F）から利用される。
func registerEmbedRoutes(g *gin.RouterGroup) {
	h := NewEmbedHandler(embed.NewFetcher())
	g.GET("/embeds/oembed", h.Resolve)
}
