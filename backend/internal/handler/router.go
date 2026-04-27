package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// NewRouter は API ルーティングを組み立てる。
// /api/v2/* は Spring Boot の /api/* と並行運用する Go 側のエンドポイント。
func NewRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "FreStyle Go backend"})
	})

	v2 := r.Group("/api/v2")

	// Phase 1: 認証不要のヘルスチェック
	healthHandler := NewHealthHandler(
		usecase.NewCheckHealthUseCase(repository.NewHealthRepository(db)),
	)
	v2.GET("/health", healthHandler.Get)

	// Phase 2: 認証 (Cognito)
	userRepo := repository.NewUserRepository(db)
	authHandler := NewAuthHandler(usecase.NewGetCurrentUserUseCase(userRepo))
	v2.POST("/auth/cognito/logout", authHandler.Logout)

	// 認証必須グループ
	authed := v2.Group("")
	authed.Use(middleware.JWTAuth())
	authed.GET("/auth/me", authHandler.Me)

	// Phase 3: AI チャット
	aiSessionRepo := repository.NewAiChatSessionRepository(db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
	)
	authed.GET("/ai-chat/sessions", aiHandler.GetSessions)
	authed.POST("/ai-chat/sessions", aiHandler.CreateSession)

	// Phase 4: ユーザー間チャット (ルーム CRUD)
	chatRoomRepo := repository.NewChatRoomRepository(db)
	chatHandler := NewChatHandler(
		usecase.NewGetChatRoomsByUserIDUseCase(chatRoomRepo),
		usecase.NewCreateChatRoomUseCase(chatRoomRepo),
	)
	authed.GET("/chat/rooms", chatHandler.GetRooms)
	authed.POST("/chat/rooms", chatHandler.CreateRoom)

	// Phase 5: プロフィール
	profileRepo := repository.NewProfileRepository(db)
	profileHandler := NewProfileHandler(
		usecase.NewGetProfileUseCase(profileRepo),
		usecase.NewUpdateProfileUseCase(profileRepo),
	)
	authed.GET("/profile/:userId", profileHandler.Get)
	authed.PUT("/profile/:userId", profileHandler.Update)

	// Phase 6: ユーザー統計
	statsHandler := NewUserStatsHandler(
		usecase.NewGetUserStatsUseCase(repository.NewUserStatsRepository(db)),
	)
	authed.GET("/user-stats/:userId", statsHandler.Get)

	// Phase 7: 練習モード (シナリオ)
	practiceRepo := repository.NewPracticeScenarioRepository(db)
	practiceHandler := NewPracticeHandler(
		usecase.NewListPracticeScenariosUseCase(practiceRepo),
		usecase.NewGetPracticeScenarioUseCase(practiceRepo),
	)
	authed.GET("/practice/scenarios", practiceHandler.List)
	authed.GET("/practice/scenarios/:id", practiceHandler.Get)

	// Phase 8: シナリオブックマーク
	bookmarkRepo := repository.NewScenarioBookmarkRepository(db)
	bookmarkHandler := NewScenarioBookmarkHandler(
		usecase.NewListScenarioBookmarksUseCase(bookmarkRepo),
		usecase.NewAddScenarioBookmarkUseCase(bookmarkRepo),
		usecase.NewRemoveScenarioBookmarkUseCase(bookmarkRepo),
	)
	authed.GET("/scenario-bookmarks", bookmarkHandler.List)
	authed.POST("/scenario-bookmarks", bookmarkHandler.Add)
	authed.DELETE("/scenario-bookmarks/:userId/:scenarioId", bookmarkHandler.Remove)

	// Phase 9: 共有 AI 会話セッション
	sharedRepo := repository.NewSharedSessionRepository(db)
	sharedHandler := NewSharedSessionHandler(
		usecase.NewListSharedSessionsUseCase(sharedRepo),
		usecase.NewCreateSharedSessionUseCase(sharedRepo),
	)
	authed.GET("/shared-sessions", sharedHandler.List)
	authed.POST("/shared-sessions", sharedHandler.Create)

	// Phase 10: Note CRUD
	noteRepo := repository.NewNoteRepository(db)
	noteHandler := NewNoteHandler(
		usecase.NewListNotesByUserIDUseCase(noteRepo),
		usecase.NewCreateNoteUseCase(noteRepo),
		usecase.NewUpdateNoteUseCase(noteRepo),
		usecase.NewDeleteNoteUseCase(noteRepo),
	)
	authed.GET("/notes", noteHandler.List)
	authed.POST("/notes", noteHandler.Create)
	authed.PUT("/notes/:id", noteHandler.Update)
	authed.DELETE("/notes/:id", noteHandler.Delete)

	// Phase 11: Note image (S3 presigned upload)
	noteImageHandler := NewNoteImageHandler(
		usecase.NewIssueNoteImageUploadURLUseCase(
			repository.NewStubNoteImagePresigner("frestyle-prod-note-images"),
		),
	)
	authed.POST("/notes/images/upload-url", noteImageHandler.IssueUploadURL)

	// Phase 12: SessionNote (セッション固有ノート)
	sessionNoteRepo := repository.NewSessionNoteRepository(db)
	sessionNoteHandler := NewSessionNoteHandler(
		usecase.NewGetSessionNoteUseCase(sessionNoteRepo),
		usecase.NewUpsertSessionNoteUseCase(sessionNoteRepo),
	)
	authed.GET("/sessions/:sessionId/note", sessionNoteHandler.Get)
	authed.PUT("/sessions/:sessionId/note", sessionNoteHandler.Upsert)

	// Phase 13: ScoreCard
	scoreCardRepo := repository.NewScoreCardRepository(db)
	scoreCardHandler := NewScoreCardHandler(
		usecase.NewListScoreCardsByUserIDUseCase(scoreCardRepo),
		usecase.NewCreateScoreCardUseCase(scoreCardRepo),
	)
	authed.GET("/score-cards", scoreCardHandler.List)
	authed.POST("/score-cards", scoreCardHandler.Create)

	// Phase 14: ScoreGoal
	scoreGoalRepo := repository.NewScoreGoalRepository(db)
	scoreGoalHandler := NewScoreGoalHandler(
		usecase.NewGetScoreGoalUseCase(scoreGoalRepo),
		usecase.NewUpsertScoreGoalUseCase(scoreGoalRepo),
	)
	authed.GET("/score-goals/:userId", scoreGoalHandler.Get)
	authed.PUT("/score-goals/:userId", scoreGoalHandler.Upsert)

	return r
}
