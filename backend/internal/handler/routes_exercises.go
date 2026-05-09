package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerExerciseRoutes は運営マスタ演習問題の閲覧 + コード実行 API を登録する。
//
// 旧 `registerPhpRoutes` を「言語非依存」に汎用化したもの:
//   - GET  /api/v2/exercises?language=php   一覧（query で言語絞り込み）
//   - GET  /api/v2/exercises/:slug          詳細（入出力例の配列を含む、 paiza 風 URL）
//   - POST /api/v2/code/execute             コード実行（body の language で言語切替）
//
// 提出履歴 / 採点 / 集計統計の API は PR-W で追加。
func registerExerciseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := repository.NewMasterExerciseRepository(deps.db)
	examplesRepo := repository.NewMasterExerciseExampleRepository(deps.db)
	exerciseHandler := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(exerciseRepo),
		usecase.NewGetMasterExerciseUseCase(exerciseRepo, examplesRepo),
	)
	g.GET("/exercises", exerciseHandler.List)
	g.GET("/exercises/:slug", exerciseHandler.GetBySlug)

	codeExecuteHandler := NewCodeExecuteHandler(usecase.NewExecuteCodeUseCase())
	g.POST("/code/execute", codeExecuteHandler.Execute)
}
