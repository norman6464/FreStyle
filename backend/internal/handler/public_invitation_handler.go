package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// PublicInvitationHandler は招待マジックリンク受諾画面用の **認証不要** エンドポイントを提供する。
// 通常 admin 系 (POST/GET/DELETE /admin/invitations) は AdminInvitationHandler が担当するが、
// 受諾画面はログイン前のユーザーが踏むため、認証必須グループの外に切り出す。
type PublicInvitationHandler struct {
	validate *usecase.ValidateInvitationTokenUseCase
}

func NewPublicInvitationHandler(v *usecase.ValidateInvitationTokenUseCase) *PublicInvitationHandler {
	return &PublicInvitationHandler{validate: v}
}

// Validate は GET /api/v2/invitations/accept/:token。
//
// 仕様:
//   - 認証不要（招待リンクをまだログインしていないユーザーが踏むため）
//   - token が空・該当なし・期限切れ・既受諾・canceled は **すべて 404** （メタ情報を漏らさない）
//   - 成功時は role / displayName / companyId / companyName のみ返す（email は返さない）
//
// レート制限はミドルウェア層で別途実装する想定（現状は未実装）。
func (h *PublicInvitationHandler) Validate(c *gin.Context) {
	token := c.Param("token")
	got, err := h.validate.Execute(c.Request.Context(), token)
	if err != nil {
		// 内部エラーは 500 だが、フロントには「招待が無効」とだけ伝えてメタ情報を出さない。
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
