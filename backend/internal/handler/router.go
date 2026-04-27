package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/config"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// NewRouter は API ルーティングを組み立てる。
// /api/v2/* は Go バックエンドの単独エンドポイント（旧 Spring Boot /api/* は廃止済み）。
func NewRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	// 全エンドポイントで CORS を許可（normanblog.com など allowlisted origin）
	r.Use(middleware.CORS())

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
	authHandler := NewAuthHandler(usecase.NewGetCurrentUserUseCase(userRepo), userRepo, &cfg.Cognito)
	v2.POST("/auth/cognito/logout", authHandler.Logout)
	// callback は code を受け取って token に交換するので認証不要
	v2.POST("/auth/cognito/callback", authHandler.Callback)
	// refresh-token は HttpOnly Cookie の refresh_token を読むため認証 middleware の対象外
	v2.POST("/auth/cognito/refresh-token", authHandler.Refresh)

	// 認証必須グループ
	authed := v2.Group("")
	authed.Use(middleware.JWTAuth())
	authed.GET("/auth/me", authHandler.Me)
	// フロントが旧 Spring Boot 互換パス /api/v2/auth/cognito/me を叩く場合の alias
	authed.GET("/auth/cognito/me", authHandler.Me)

	// Phase 3: AI チャット
	aiSessionRepo := repository.NewAiChatSessionRepository(db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
	)
	authed.GET("/ai-chat/sessions", aiHandler.GetSessions)
	authed.POST("/ai-chat/sessions", aiHandler.CreateSession)
	// 旧 Spring Boot 互換 path: フロントは /chat/ai/sessions を叩く
	authed.GET("/chat/ai/sessions", aiHandler.GetSessions)
	authed.POST("/chat/ai/sessions", aiHandler.CreateSession)

	// Phase 4: ユーザー間チャット (ルーム CRUD)
	chatRoomRepo := repository.NewChatRoomRepository(db)
	chatHandler := NewChatHandler(
		usecase.NewGetChatRoomsByUserIDUseCase(chatRoomRepo),
		usecase.NewCreateChatRoomUseCase(chatRoomRepo),
	)
	authed.GET("/chat/rooms", chatHandler.GetRooms)
	authed.POST("/chat/rooms", chatHandler.CreateRoom)
	// /chat/stats はフロント MenuPage が叩いていた旧 Spring Boot path。
	// Go では未実装なので暫定 stub で空オブジェクトを返す（フロントは PR で chat/rooms 由来に切替済）。
	authed.GET("/chat/stats", func(c *gin.Context) {
		c.JSON(200, gin.H{"chatPartnerCount": 0})
	})

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
	// 旧 Spring Boot 互換 path のための alias（フロントが /api/v2/bookmarks を叩くケース）
	authed.GET("/bookmarks", bookmarkHandler.List)

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
	// 旧 Spring Boot 互換 path の alias
	authed.GET("/scores/history", scoreCardHandler.List)

	// Phase 14: ScoreGoal
	scoreGoalRepo := repository.NewScoreGoalRepository(db)
	scoreGoalHandler := NewScoreGoalHandler(
		usecase.NewGetScoreGoalUseCase(scoreGoalRepo),
		usecase.NewUpsertScoreGoalUseCase(scoreGoalRepo),
	)
	authed.GET("/score-goals/:userId", scoreGoalHandler.Get)
	authed.PUT("/score-goals/:userId", scoreGoalHandler.Upsert)
	// 旧 Spring Boot 互換 path（単数形）: フロントが /score-goal を叩く
	authed.GET("/score-goal", func(c *gin.Context) {
		c.JSON(200, gin.H{"targetScore": 0})
	})

	// Phase 15: ScoreTrend
	scoreTrendHandler := NewScoreTrendHandler(
		usecase.NewGetScoreTrendUseCase(repository.NewScoreTrendRepository(db)),
	)
	authed.GET("/score-trends/:userId", scoreTrendHandler.Get)

	// Phase 16: Ranking
	rankingHandler := NewRankingHandler(
		usecase.NewGetRankingUseCase(repository.NewRankingRepository(db)),
	)
	authed.GET("/rankings", rankingHandler.Get)

	// Phase 17: LearningReport (非同期生成)
	learningReportRepo := repository.NewLearningReportRepository(db)
	learningReportHandler := NewLearningReportHandler(
		usecase.NewListLearningReportsUseCase(learningReportRepo),
		usecase.NewRequestLearningReportUseCase(learningReportRepo, repository.NewStubSqsEnqueuer()),
	)
	authed.GET("/learning-reports", learningReportHandler.List)
	authed.POST("/learning-reports", learningReportHandler.Request)

	// Phase 18: ConversationTemplate
	templateHandler := NewConversationTemplateHandler(
		usecase.NewListConversationTemplatesUseCase(repository.NewConversationTemplateRepository(db)),
	)
	authed.GET("/conversation-templates", templateHandler.List)

	// Phase 19: FavoritePhrase
	favRepo := repository.NewFavoritePhraseRepository(db)
	favHandler := NewFavoritePhraseHandler(
		usecase.NewListFavoritePhrasesUseCase(favRepo),
		usecase.NewAddFavoritePhraseUseCase(favRepo),
		usecase.NewDeleteFavoritePhraseUseCase(favRepo),
	)
	authed.GET("/favorite-phrases", favHandler.List)
	authed.POST("/favorite-phrases", favHandler.Add)
	authed.DELETE("/favorite-phrases/:id", favHandler.Remove)

	// Phase 20: Friendship
	friendshipRepo := repository.NewFriendshipRepository(db)
	friendshipHandler := NewFriendshipHandler(
		usecase.NewListFriendshipsUseCase(friendshipRepo),
		usecase.NewRequestFriendshipUseCase(friendshipRepo),
		usecase.NewRespondFriendshipUseCase(friendshipRepo),
	)
	authed.GET("/friendships", friendshipHandler.List)
	authed.POST("/friendships", friendshipHandler.Request)
	authed.PATCH("/friendships/:id", friendshipHandler.Respond)

	// Phase 21: Notification
	notificationRepo := repository.NewNotificationRepository(db)
	notificationHandler := NewNotificationHandler(
		usecase.NewListNotificationsUseCase(notificationRepo),
		usecase.NewMarkNotificationReadUseCase(notificationRepo),
	)
	authed.GET("/notifications", notificationHandler.List)
	authed.PATCH("/notifications/:id/read", notificationHandler.MarkRead)

	// Phase 22: ReminderSetting
	reminderRepo := repository.NewReminderSettingRepository(db)
	reminderHandler := NewReminderSettingHandler(
		usecase.NewGetReminderSettingUseCase(reminderRepo),
		usecase.NewUpsertReminderSettingUseCase(reminderRepo),
	)
	authed.GET("/reminder-settings/:userId", reminderHandler.Get)
	authed.PUT("/reminder-settings/:userId", reminderHandler.Upsert)

	// Phase 23: DailyGoal
	dailyGoalRepo := repository.NewDailyGoalRepository(db)
	dailyGoalHandler := NewDailyGoalHandler(
		usecase.NewGetDailyGoalUseCase(dailyGoalRepo),
		usecase.NewUpsertDailyGoalUseCase(dailyGoalRepo),
	)
	authed.GET("/daily-goals/:userId", dailyGoalHandler.Get)
	authed.PUT("/daily-goals/:userId", dailyGoalHandler.Upsert)
	// /daily-goals/today は Spring Boot 時代の path。Go では :userId 解析されて 400 になっていた。
	// 暫定で 200 + 空オブジェクトを返す stub。本来は current user の今日の goal を返すべきだが別 issue。
	authed.GET("/daily-goals/today", func(c *gin.Context) {
		c.JSON(200, gin.H{"date": "", "targetMinutes": 0, "actualMinutes": 0, "isAchieved": false})
	})

	// Phase 24: WeeklyChallenge
	weeklyRepo := repository.NewWeeklyChallengeRepository(db)
	weeklyHandler := NewWeeklyChallengeHandler(
		usecase.NewGetCurrentWeeklyChallengeUseCase(weeklyRepo),
		usecase.NewCompleteWeeklyChallengeUseCase(weeklyRepo),
	)
	authed.GET("/weekly-challenges/current", weeklyHandler.GetCurrent)
	authed.POST("/weekly-challenges/complete", weeklyHandler.Complete)

	// Phase 25: AdminInvitation
	adminInvRepo := repository.NewAdminInvitationRepository(db)
	adminInvHandler := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(adminInvRepo),
		usecase.NewCreateAdminInvitationUseCase(adminInvRepo, repository.NewStubCognitoAdminClient()),
		usecase.NewCancelAdminInvitationUseCase(adminInvRepo),
	)
	authed.GET("/admin/invitations", adminInvHandler.List)
	authed.POST("/admin/invitations", adminInvHandler.Create)
	authed.DELETE("/admin/invitations/:id", adminInvHandler.Cancel)

	// Phase 26: AdminScenario
	adminScenarioRepo := repository.NewAdminScenarioRepository(db)
	adminScenarioHandler := NewAdminScenarioHandler(
		usecase.NewListAdminScenariosUseCase(adminScenarioRepo),
		usecase.NewCreateAdminScenarioUseCase(adminScenarioRepo),
		usecase.NewUpdateAdminScenarioUseCase(adminScenarioRepo),
		usecase.NewDeleteAdminScenarioUseCase(adminScenarioRepo),
	)
	authed.GET("/admin/scenarios", adminScenarioHandler.List)
	authed.POST("/admin/scenarios", adminScenarioHandler.Create)
	authed.PUT("/admin/scenarios/:id", adminScenarioHandler.Update)
	authed.DELETE("/admin/scenarios/:id", adminScenarioHandler.Delete)

	// Phase 27: AiChat WebSocket (echo skeleton, Bedrock 連携は Phase 27.1)
	aiChatWsHandler := NewAiChatWsHandler()
	authed.GET("/ws/ai-chat", aiChatWsHandler.Handle)
	// SockJS 互換 info endpoint。フロント (@stomp/stompjs + sockjs-client) は接続前に
	// /ws/.../info?t=... を叩く。Go は raw WebSocket なので websocket=false で返して
	// SockJS の reconnect ループを早期終了させる。本格的な WS 統合は別 issue。
	v2.GET("/ws/ai-chat/info", sockJSInfo)
	v2.GET("/ws/chat/info", sockJSInfo)
	// /ws/... (api/v2 プレフィックス無し) でも叩かれているので root にも登録
	r.GET("/ws/ai-chat/info", sockJSInfo)
	r.GET("/ws/chat/info", sockJSInfo)

	// Phase 28: Chat WebSocket (ルームごとブロードキャスト)
	chatWsHandler := NewChatWsHandler()
	authed.GET("/ws/chat/:roomId", chatWsHandler.Handle)

	return r
}

// sockJSInfo は SockJS 互換の /info レスポンスを返す。
// websocket=false を返すことで sockjs-client が WS 試行をスキップし、
// 全 transport 失敗 → 早期に諦めるようになる（本格的な WS 統合は別 issue）。
func sockJSInfo(c *gin.Context) {
	c.JSON(200, gin.H{
		"entropy":       0,
		"origins":       []string{"*:*"},
		"cookie_needed": false,
		"websocket":     false,
	})
}
