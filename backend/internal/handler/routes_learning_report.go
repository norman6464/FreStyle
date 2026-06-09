package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerLearningReportRoutes は月次学習レポートの取得 / 生成要求を登録する。
// 生成は SQS 非同期ジョブで、当面は stub enqueuer を使う。
func registerLearningReportRoutes(g *gin.RouterGroup, deps *routeDeps) {
	repo := persistence.NewLearningReportRepository(deps.db)
	queue := persistence.NewStubSqsEnqueuer()
	h := NewLearningReportHandler(
		usecase.NewListLearningReportsUseCase(repo),
		usecase.NewRequestLearningReportUseCase(repo, queue),
	)
	g.GET("/learning-reports", h.List)
	g.POST("/learning-reports/generate", h.Request)
}
