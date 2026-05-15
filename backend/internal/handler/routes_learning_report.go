package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerLearningReportRoutes は月次学習レポートの取得 / 生成要求を登録する。
//
//	GET  /api/v2/learning-reports             current user のレポート一覧
//	POST /api/v2/learning-reports/generate    `{year, month}` で月次レポート生成を要求
//
// 生成は SQS にキューする非同期ジョブで、 当面は stub enqueuer を使用する
// （実 SQS 接続は別 PR / 別 infra で導入予定）。
func registerLearningReportRoutes(g *gin.RouterGroup, deps *routeDeps) {
	repo := legacyrepository.NewLearningReportRepository(deps.db)
	queue := legacyrepository.NewStubSqsEnqueuer()
	h := NewLearningReportHandler(
		usecase.NewListLearningReportsUseCase(repo),
		usecase.NewRequestLearningReportUseCase(repo, queue),
	)
	g.GET("/learning-reports", h.List)
	g.POST("/learning-reports/generate", h.Request)
}
