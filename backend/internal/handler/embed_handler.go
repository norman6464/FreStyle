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

// Resolve は ?url= の OGP / oEmbed を解決して返す。
// 失敗は InvalidURL→400 / UnsupportedHost→422 / Unreachable→502 / その他→500 にマップする。
//
//	@Summary      外部 URL の OGP / oEmbed メタ 取得
//	@Description  ?url= で 指定 した URL の OGP / oEmbed を 解決 し card 形式 で 返す。 allow-list / SSRF / DNS rebinding 対策 は infra/embed.Fetcher が 担う。
//	@Tags         embeds
//	@Produce      json
//	@Param        url  query     string  true  "メタ 取得 対象 URL"
//	@Success      200  {object}  github_com_norman6464_FreStyle_backend_internal_infra_embed.Card
//	@Failure      400  {object}  errorResponse  "url 欠落 or 無効 URL"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      422  {object}  errorResponse  "allow-list 外 ホスト"
//	@Failure      502  {object}  errorResponse  "到達 不可"
//	@Failure      500  {object}  errorResponse  "内部 エラー"
//	@Router       /embeds/oembed [get]
//	@Security     CookieAuth
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
