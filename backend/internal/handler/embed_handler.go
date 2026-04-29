package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/infra/embed"
)

// EmbedHandler は外部 URL のメタ情報 (OGP / oEmbed) を取得して返す。
// SSRF / DNS rebinding 対策は infra/embed.Fetcher 内で完結している。
type EmbedHandler struct {
	fetcher *embed.Fetcher
}

func NewEmbedHandler(f *embed.Fetcher) *EmbedHandler {
	return &EmbedHandler{fetcher: f}
}

// Resolve は GET /api/v2/embeds/oembed?url=...
//
// ?url= が必須クエリパラメータ。fetcher が allow-list / SSRF ガードを担う。
// 失敗時のステータスマップ:
//   - ErrInvalidURL      → 400
//   - ErrUnsupportedHost → 422
//   - ErrUnreachable     → 502
//   - その他             → 500
func (h *EmbedHandler) Resolve(c *gin.Context) {
	raw := c.Query("url")
	if raw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}
	card, err := h.fetcher.Resolve(c.Request.Context(), raw)
	if err != nil {
		switch {
		case errors.Is(err, embed.ErrInvalidURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_url"})
		case errors.Is(err, embed.ErrUnsupportedHost):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unsupported_host"})
		case errors.Is(err, embed.ErrUnreachable):
			c.JSON(http.StatusBadGateway, gin.H{"error": "unreachable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.JSON(http.StatusOK, card)
}
