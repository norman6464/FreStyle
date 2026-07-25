package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerDailyGoalsRoutes は日次学習目標の API を登録する。
func registerDailyGoalsRoutes(g *gin.RouterGroup, deps *routeDeps) {
	activityRepo := persistence.NewUserDailyActivityRepository(deps.db)
	h := NewDailyGoalsHandler(usecase.NewGetDailyStreakUseCase(activityRepo))
	g.GET("/daily-goals/streak", h.GetStreak)
}
