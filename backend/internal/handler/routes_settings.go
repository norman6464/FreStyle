package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerSettingsRoutes はリマインダー設定と日次目標の REST エンドポイントを登録する。
func registerSettingsRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 22: ReminderSetting
	reminderRepo := repository.NewReminderSettingRepository(deps.db)
	reminderHandler := NewReminderSettingHandler(
		usecase.NewGetReminderSettingUseCase(reminderRepo),
		usecase.NewUpsertReminderSettingUseCase(reminderRepo),
	)
	g.GET("/reminder-settings/:userId", reminderHandler.Get)
	g.PUT("/reminder-settings/:userId", reminderHandler.Upsert)

	// Phase 23: DailyGoal
	dailyGoalRepo := repository.NewDailyGoalRepository(deps.db)
	dailyGoalHandler := NewDailyGoalHandler(
		usecase.NewGetDailyGoalUseCase(dailyGoalRepo),
		usecase.NewUpsertDailyGoalUseCase(dailyGoalRepo),
	)
	// /daily-goals/today はフロント MenuPage が呼ぶ。handler 側で path の :userId が "today" や数値以外なら
	// current user に解決する仕組みを入れている (see daily_goal_handler.go resolveUserID)。
	g.GET("/daily-goals/:userId", dailyGoalHandler.Get)
	g.PUT("/daily-goals/:userId", dailyGoalHandler.Upsert)
}
