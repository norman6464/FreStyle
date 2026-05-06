package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// OnboardingHandler は Welcome 画面の完了通知を受ける認証必須エンドポイント。
// 別途 ProfileHandler に同居させる選択肢もあったが、責務（onboarding 状態管理）が
// 明確に独立しているので別 handler に切り出す（将来「再オンボーディング」「skipped」
// などのバリエーションが増えても影響範囲を局所化できる）。
type OnboardingHandler struct {
	complete *usecase.CompleteOnboardingUseCase
}

func NewOnboardingHandler(complete *usecase.CompleteOnboardingUseCase) *OnboardingHandler {
	return &OnboardingHandler{complete: complete}
}

// Complete は POST /api/v2/profile/me/onboarding/complete。
// 自分自身の users.onboarded_at を NOW() に更新する。
// repo 側で IS NULL ガード付きなので冪等（複数回呼んでも初回日時が保持される）。
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
