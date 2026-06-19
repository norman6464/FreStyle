package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerLessonProgressRoutes は trainee 自身の学習進捗（レッスン完了状態）の API を登録する。
// すべて current user 名義（userId は受け取らない）。
func registerLessonProgressRoutes(g *gin.RouterGroup, deps *routeDeps) {
	progressRepo := persistence.NewLessonProgressRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)
	courseRepo := persistence.NewCourseRepository(deps.db)
	activityRepo := persistence.NewUserDailyActivityRepository(deps.db)
	h := NewLessonProgressHandler(
		usecase.NewMarkLessonCompletedUseCase(progressRepo, materialRepo, courseRepo, activityRepo),
		usecase.NewMarkLessonIncompleteUseCase(progressRepo),
		usecase.NewListLessonProgressUseCase(progressRepo),
	)
	g.GET("/lesson-progress", h.List)
	g.POST("/lesson-progress", h.Complete)
	g.DELETE("/lesson-progress/:teachingMaterialId", h.Incomplete)
}
