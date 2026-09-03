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
// 成功時は role / name / workspaceName / workspaceId のみ返す（email は返さない）。
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
	resp := gin.H{
		"role":          got.Role,
		"name":          got.Name,
		"workspaceName": got.WorkspaceName,
	}
	if got.WorkspaceID != nil {
		resp["workspaceId"] = got.WorkspaceID
	}
	c.JSON(http.StatusOK, resp)
}
