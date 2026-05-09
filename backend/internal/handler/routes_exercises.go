package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerExerciseRoutes は運営マスタ演習問題の閲覧 + 提出 + 採点 API を登録する。
//
// 旧 `registerPhpRoutes` を「言語非依存」に汎用化したもの:
//   - GET  /api/v2/exercises?language=php          一覧（current user 状態 + 全体集計を含む）
//   - GET  /api/v2/exercises/:slug                 詳細（入出力例の配列を含む、 slug ベース URL）
//   - POST /api/v2/exercises/:slug/submit          コードを提出して採点（履歴に保存）
//   - GET  /api/v2/exercises/:slug/submissions     current user の履歴
//   - POST /api/v2/code/execute                    コード実行（採点せず stdout だけ返す）
func registerExerciseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := repository.NewMasterExerciseRepository(deps.db)
	examplesRepo := repository.NewMasterExerciseExampleRepository(deps.db)
	submissionsRepo := repository.NewExerciseSubmissionRepository(deps.db)
	executor := usecase.NewExecuteCodeUseCase()

	exerciseHandler := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(exerciseRepo),
		usecase.NewListMasterExercisesWithStatusUseCase(exerciseRepo, submissionsRepo),
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
