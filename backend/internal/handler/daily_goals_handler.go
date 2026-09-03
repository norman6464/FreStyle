package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// DailyGoalsHandler は日次学習目標まわりの API を提供する。
type DailyGoalsHandler struct {
	getStreak *usecase.GetDailyStreakUseCase
}

func NewDailyGoalsHandler(s *usecase.GetDailyStreakUseCase) *DailyGoalsHandler {
	return &DailyGoalsHandler{getStreak: s}
}

// GetStreak はログイン中ユーザーの連続学習日数の統計を返す。
func (h *DailyGoalsHandler) GetStreak(c *gin.Context) {
	userID := middleware.CurrentUserIDOrZero(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	out, err := h.getStreak.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "学習統計の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, out)
}
