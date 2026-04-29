package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerPracticeRoutes は練習モード（シナリオ・ブックマーク・共有セッション）の
// エンドポイントを登録する。
func registerPracticeRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 7: 練習モード (シナリオ)
	practiceRepo := repository.NewPracticeScenarioRepository(deps.db)
	practiceHandler := NewPracticeHandler(
		usecase.NewListPracticeScenariosUseCase(practiceRepo),
		usecase.NewGetPracticeScenarioUseCase(practiceRepo),
	)
	g.GET("/practice/scenarios", practiceHandler.List)
	g.GET("/practice/scenarios/:id", practiceHandler.Get)

	// Phase 8: シナリオブックマーク
	bookmarkRepo := repository.NewScenarioBookmarkRepository(deps.db)
	bookmarkHandler := NewScenarioBookmarkHandler(
		usecase.NewListScenarioBookmarksUseCase(bookmarkRepo),
		usecase.NewAddScenarioBookmarkUseCase(bookmarkRepo),
		usecase.NewRemoveScenarioBookmarkUseCase(bookmarkRepo),
	)
	// scenario-bookmarks は current user の所有物のみ操作可能（IDOR 対策で userId は受けない）。
	// フロントは /scenario-bookmarks/:scenarioId に POST/DELETE を投げる。
	g.GET("/scenario-bookmarks", bookmarkHandler.List)
	g.POST("/scenario-bookmarks/:scenarioId", bookmarkHandler.Add)
	g.DELETE("/scenario-bookmarks/:scenarioId", bookmarkHandler.Remove)

	// Phase 9: 共有 AI 会話セッション
	sharedRepo := repository.NewSharedSessionRepository(deps.db)
	sharedHandler := NewSharedSessionHandler(
		usecase.NewListSharedSessionsUseCase(sharedRepo),
		usecase.NewCreateSharedSessionUseCase(sharedRepo),
	)
	g.GET("/shared-sessions", sharedHandler.List)
	g.POST("/shared-sessions", sharedHandler.Create)
}
