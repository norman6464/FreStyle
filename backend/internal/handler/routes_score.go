package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerScoreRoutes はスコアカード・スコアゴール・トレンド・ランキング・
// 学習レポート・週次チャレンジを含む「スコア / 進捗」ドメインのエンドポイントをまとめて登録する。
func registerScoreRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 13: ScoreCard
	scoreCardRepo := repository.NewScoreCardRepository(deps.db)
	scoreCardHandler := NewScoreCardHandler(
		usecase.NewListScoreCardsByUserIDUseCase(scoreCardRepo),
		usecase.NewCreateScoreCardUseCase(scoreCardRepo),
	)
	g.GET("/score-cards", scoreCardHandler.List)
	g.POST("/score-cards", scoreCardHandler.Create)

	// Phase 14: ScoreGoal
	scoreGoalRepo := repository.NewScoreGoalRepository(deps.db)
	scoreGoalHandler := NewScoreGoalHandler(
		usecase.NewGetScoreGoalUseCase(scoreGoalRepo),
		usecase.NewUpsertScoreGoalUseCase(scoreGoalRepo),
	)
	// 認証済 current user の score goal のみを返す。
	// 旧 `/score-goals/:userId` 系は admin 認可機構が無く IDOR になるため廃止。
	// admin 用の他ユーザー閲覧/変更が必要になったら別途 /admin/users/:id/score-goal として追加する。
	g.GET("/score-goals", scoreGoalHandler.Get)
	g.PUT("/score-goals", scoreGoalHandler.Upsert)

	// Phase 15: ScoreTrend
	scoreTrendHandler := NewScoreTrendHandler(
		usecase.NewGetScoreTrendUseCase(repository.NewScoreTrendRepository(deps.db)),
	)
	g.GET("/score-trends/:userId", scoreTrendHandler.Get)

	// Phase 16: Ranking
	rankingHandler := NewRankingHandler(
		usecase.NewGetRankingUseCase(repository.NewRankingRepository(deps.db)),
	)
	g.GET("/rankings", rankingHandler.Get)

	// Phase 17: LearningReport (非同期生成)
	learningReportRepo := repository.NewLearningReportRepository(deps.db)
	learningReportHandler := NewLearningReportHandler(
		usecase.NewListLearningReportsUseCase(learningReportRepo),
		usecase.NewRequestLearningReportUseCase(learningReportRepo, repository.NewStubSqsEnqueuer()),
	)
	g.GET("/learning-reports", learningReportHandler.List)
	g.POST("/learning-reports", learningReportHandler.Request)

	// Phase 24: WeeklyChallenge
	weeklyRepo := repository.NewWeeklyChallengeRepository(deps.db)
	weeklyHandler := NewWeeklyChallengeHandler(
		usecase.NewGetCurrentWeeklyChallengeUseCase(weeklyRepo),
		usecase.NewCompleteWeeklyChallengeUseCase(weeklyRepo),
	)
	g.GET("/weekly-challenges/current", weeklyHandler.GetCurrent)
	g.POST("/weekly-challenges/complete", weeklyHandler.Complete)
}
