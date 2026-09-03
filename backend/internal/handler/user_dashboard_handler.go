package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// UserDashboardHandler はパーソナライズダッシュボードデータを返す。
type UserDashboardHandler struct {
	getDashboard *usecase.GetUserDashboardUseCase
}

func NewUserDashboardHandler(d *usecase.GetUserDashboardUseCase) *UserDashboardHandler {
	return &UserDashboardHandler{getDashboard: d}
}

// Get はログイン中ユーザーのダッシュボードデータ（streak / 活動カレンダー / 「続きから」）を返す。
func (h *UserDashboardHandler) Get(c *gin.Context) {
	userID := middleware.CurrentUserIDOrZero(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	out, err := h.getDashboard.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ダッシュボードの取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, out)
}
