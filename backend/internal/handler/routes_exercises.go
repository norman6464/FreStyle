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
//   - GET  /api/v2/exercises/:id            個別取得
//   - POST /api/v2/code/execute             コード実行（body の language で言語切替）
//
// 会社作成の company_exercises / 提出履歴の exercise_submissions 用 API は
// 別 PR (PR-H1 以降) で追加する。
func registerExerciseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := repository.NewMasterExerciseRepository(deps.db)
	exerciseHandler := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(exerciseRepo),
		usecase.NewGetMasterExerciseUseCase(exerciseRepo),
	)
	g.GET("/exercises", exerciseHandler.List)
	g.GET("/exercises/:id", exerciseHandler.Get)

	codeExecuteHandler := NewCodeExecuteHandler(usecase.NewExecuteCodeUseCase())
	g.POST("/code/execute", codeExecuteHandler.Execute)
}
