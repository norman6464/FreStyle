package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// OnboardingHandler は Welcome 画面の完了通知を受ける認証必須エンドポイント。
type OnboardingHandler struct {
	complete *usecase.CompleteOnboardingUseCase
}

func NewOnboardingHandler(complete *usecase.CompleteOnboardingUseCase) *OnboardingHandler {
	return &OnboardingHandler{complete: complete}
}

// Complete は自分自身の users.onboarded_at を更新する（IS NULL ガードで冪等）。
//
//	@Summary      オンボーディング 完了 通知
//	@Description  Welcome 画面 「はじめる」 押下 で onboarded_at = NOW() に 更新。 冪等 (再 押下 で も 初回 日時 保持)。
//	@Tags         profile
//	@Success      204  "成功 (本文 なし)"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      500  {object}  errorResponse  "内部 エラー"
//	@Router       /profile/me/onboarding/complete [post]
//	@Security     CookieAuth
func (h *OnboardingHandler) Complete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.complete.Execute(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
