package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerExerciseRoutes は運営マスタ演習問題の閲覧 + 提出 + 採点 + コード実行 API を登録する。
func registerExerciseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := persistence.NewMasterExerciseRepository(deps.db)
	examplesRepo := persistence.NewMasterExerciseExampleRepository(deps.db)
	submissionsRepo := persistence.NewExerciseSubmissionRepository(deps.db)
	executor := usecase.NewExecuteCodeUseCase()

	exerciseHandler := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(exerciseRepo),
		usecase.NewListMasterExercisesWithStatusUseCase(exerciseRepo),
		usecase.NewGetMasterExerciseUseCase(exerciseRepo, examplesRepo),
	)
	g.GET("/exercises", exerciseHandler.List)
	g.GET("/exercises/:slug", exerciseHandler.GetBySlug)

	submissionHandler := NewExerciseSubmissionHandler(
		usecase.NewSubmitMasterExerciseUseCase(exerciseRepo, examplesRepo, submissionsRepo, executor),
		usecase.NewListUserMasterSubmissionsUseCase(exerciseRepo, submissionsRepo),
	)
	g.POST("/exercises/:slug/submit", submissionHandler.Submit)
	g.GET("/exercises/:slug/submissions", submissionHandler.List)

	codeExecuteHandler := NewCodeExecuteHandler(executor)
	g.POST("/code/execute", codeExecuteHandler.Execute)
}
