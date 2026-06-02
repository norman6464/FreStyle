package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// PublicInvitationHandler は招待受諾画面用の認証不要エンドポイントを提供する
// （ログイン前のユーザーが踏むため認証必須グループの外に置く）。
type PublicInvitationHandler struct {
	validate *usecase.ValidateInvitationTokenUseCase
}

func NewPublicInvitationHandler(v *usecase.ValidateInvitationTokenUseCase) *PublicInvitationHandler {
	return &PublicInvitationHandler{validate: v}
}

// Validate は招待 token を検証する。無効・期限切れ・既受諾は全て 404（メタ情報を漏らさない）。
// 成功時は role / displayName / companyId / companyName のみ返す（email は返さない）。
//
//	@Summary      招待 トークン 検証 (公開 / 認証 不要)
//	@Description  招待 マジック リンク から 来た 未 ログイン user 用。 該当 / 期限 切れ / 既 受諾 等 は メタ 情報 を 漏らさず 全て 404。
//	@Tags         invitations
//	@Produce      json
//	@Param        token  path      string  true  "招待 受諾 token (UUID)"
//	@Success      200    {object}  invitationValidateResponse
//	@Failure      404    {object}  errorResponse  "招待 が 無効 / 期限 切れ"
//	@Failure      500    {object}  errorResponse  "内部 エラー"
//	@Failure      429    {object}  errorResponse  "レート制限超過"
//	@Header       429  {string}  Retry-After  "再試行までの秒数 (例: 60)"
//	@Router       /invitations/accept/{token} [get]
func (h *PublicInvitationHandler) Validate(c *gin.Context) {
	token := c.Param("token")
	got, err := h.validate.Execute(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	if got == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "invitation_not_found_or_expired"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"role":        got.Role,
		"displayName": got.DisplayName,
		"companyId":   got.CompanyID,
		"companyName": got.CompanyName,
	})
}
