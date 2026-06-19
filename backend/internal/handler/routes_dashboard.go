package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerDashboardRoutes はパーソナライズダッシュボードの API を登録する。
func registerDashboardRoutes(g *gin.RouterGroup, deps *routeDeps) {
	activityRepo := persistence.NewUserDailyActivityRepository(deps.db)
	chapterViewRepo := persistence.NewUserChapterViewRepository(deps.db)
	h := NewUserDashboardHandler(
		usecase.NewGetUserDashboardUseCase(activityRepo, chapterViewRepo),
	)
	g.GET("/me/dashboard", h.Get)
}
