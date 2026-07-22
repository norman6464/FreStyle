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
	activityRepo := persistence.NewUserDailyActivityRepository(deps.db)

	// CODE_RUNNER_URL がセットされていれば別コンテナ（サイドカー）の code-runner へ HTTP 委譲、
	// 未設定なら同プロセス内でサンドボックス実行する（ローカル / 単一イメージ運用）。
	runner := codeRunner(deps.cfg.CodeRunnerURL)
	executor := usecase.NewExecuteCodeUseCase(runner)
	warmup := usecase.NewWarmupCodeUseCase(runner)

	exerciseHandler := NewMasterExerciseHandler(
		usecase.NewListMasterExercisesUseCase(exerciseRepo),
		usecase.NewListMasterExercisesWithStatusUseCase(exerciseRepo),
		usecase.NewGetMasterExerciseUseCase(exerciseRepo, examplesRepo),
		usecase.NewGetExerciseLanguageSummaryUseCase(exerciseRepo),
	)
	g.GET("/exercises", exerciseHandler.List)
	// 静的セグメントは :slug より優先して解決される（gin v1.12 で動作確認済）。
	g.GET("/exercises/summary", exerciseHandler.Summary)
	g.GET("/exercises/:slug", exerciseHandler.GetBySlug)

	submissionHandler := NewExerciseSubmissionHandler(
		usecase.NewSubmitMasterExerciseUseCase(exerciseRepo, examplesRepo, submissionsRepo, executor, activityRepo),
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
