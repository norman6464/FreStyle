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
//
//	@Summary      連続 学習 日数 統計 取得
//	@Description  現在 の 連続 学習 日数 / 最長 連続 日数 / 累計 学習 日数 を 返す。 設定 画面 の プロフィール 統計 用。
//	@Tags         daily-goals
//	@Produce      json
//	@Success      200  {object}  usecase.GetDailyStreakOutput
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      500  {object}  errorResponse  "集計失敗"
//	@Router       /daily-goals/streak [get]
//	@Security     CookieAuth
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
