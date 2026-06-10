package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/infra/coderunner"
	"github.com/norman6464/FreStyle/backend/internal/infra/sandbox"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerExerciseRoutes は運営マスタ演習問題の閲覧 + 提出 + 採点 + コード実行 API を登録する。
func registerExerciseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := persistence.NewMasterExerciseRepository(deps.db)
	examplesRepo := persistence.NewMasterExerciseExampleRepository(deps.db)
	submissionsRepo := persistence.NewExerciseSubmissionRepository(deps.db)

	// CODE_RUNNER_URL がセットされていれば別コンテナ（サイドカー）の code-runner へ HTTP 委譲、
	// 未設定なら同プロセス内でサンドボックス実行する（ローカル / 単一イメージ運用）。
	runner := codeRunner(deps.cfg.CodeRunnerURL)
	executor := usecase.NewExecuteCodeUseCase(runner)
	warmup := usecase.NewWarmupCodeUseCase(runner)

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

	codeExecuteHandler := NewCodeExecuteHandler(executor, warmup)
	g.POST("/code/execute", codeExecuteHandler.Execute)
	g.POST("/code/warmup", codeExecuteHandler.Warmup)
}

// codeRunner は CODE_RUNNER_URL の有無で実行系（HTTP サイドカー / in-process）を選ぶ。
func codeRunner(url string) usecase.CodeRunner {
	if url != "" {
		return coderunner.NewClient(url)
	}
	return sandbox.NewRunner()
}
