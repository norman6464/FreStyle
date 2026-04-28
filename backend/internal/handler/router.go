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

	// 認証必須グループ。JWTAuth で sub を context に詰めた後、CurrentUser で users.id を解決する。
	authed := v2.Group("")
	authed.Use(middleware.JWTAuth())
	authed.Use(middleware.CurrentUser(userRepo))
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
	// /profile/:userId は数字 / "me" の両方を受ける（handler.resolveUserID で current user 解決）。
	// フロント互換のため /profile/:userId/update PUT / /profile/:userId/image/presigned-url POST も提供する。
	authed.GET("/profile/:userId", profileHandler.Get)
	authed.PUT("/profile/:userId", profileHandler.Update)
	authed.PUT("/profile/:userId/update", profileHandler.Update)

	// Profile アイコン画像の S3 presigned-url。bucket / CDN は note image と同じインフラを共有する
	// （別 prefix `profiles/` で分離）。AWS SDK 統合は別 issue。
	profileImageHandler := NewProfileImageHandler(
		usecase.NewIssueProfileImageUploadURLUseCase(
			repository.NewStubProfileImagePresigner("frestyle-prod-note-images", "https://normanblog.com"),
		),
	)
	authed.POST("/profile/:userId/image/presigned-url", profileImageHandler.IssueUploadURL)

	// Phase 6: ユーザー統計
	statsHandler := NewUserStatsHandler(
		usecase.NewGetUserStatsUseCase(repository.NewUserStatsRepository(db)),
	)
	// 旧 path /user-stats/:userId と Spring 風 /users/:userId/stats の両方を提供する。
	authed.GET("/user-stats/:userId", statsHandler.Get)
	authed.GET("/users/:userId/stats", statsHandler.Get)

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
	// scenario-bookmarks は current user の所有物のみ操作可能（IDOR 対策で userId は受けない）。
	// フロントは /scenario-bookmarks/:scenarioId に POST/DELETE を投げる。
	authed.GET("/scenario-bookmarks", bookmarkHandler.List)
	authed.POST("/scenario-bookmarks/:scenarioId", bookmarkHandler.Add)
	authed.DELETE("/scenario-bookmarks/:scenarioId", bookmarkHandler.Remove)

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
	// 認証済 current user の score goal のみを返す。
	// 旧 `/score-goals/:userId` 系は admin 認可機構が無く IDOR になるため廃止。
	// admin 用の他ユーザー閲覧/変更が必要になったら別途 /admin/users/:id/score-goal として追加する。
	authed.GET("/score-goals", scoreGoalHandler.Get)
	authed.PUT("/score-goals", scoreGoalHandler.Upsert)

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

	// Phase 20: Friendship + 単方向フォロー（フロントの follow / unfollow / status / following / followers に対応）
	friendshipRepo := repository.NewFriendshipRepository(db)
	friendshipHandler := NewFriendshipHandler(
		usecase.NewListFriendshipsUseCase(friendshipRepo),
		usecase.NewRequestFriendshipUseCase(friendshipRepo),
		usecase.NewRespondFriendshipUseCase(friendshipRepo),
		usecase.NewFollowUserUseCase(friendshipRepo),
		usecase.NewUnfollowUserUseCase(friendshipRepo),
		usecase.NewListFollowingUseCase(friendshipRepo),
		usecase.NewListFollowersUseCase(friendshipRepo),
		usecase.NewGetFollowStatusUseCase(friendshipRepo),
	)
	authed.GET("/friendships", friendshipHandler.List)
	authed.POST("/friendships", friendshipHandler.Request)
	authed.PATCH("/friendships/:id", friendshipHandler.Respond)
	authed.GET("/friendships/following", friendshipHandler.Following)
	authed.GET("/friendships/followers", friendshipHandler.Followers)
	authed.POST("/friendships/:userId/follow", friendshipHandler.Follow)
	authed.DELETE("/friendships/:userId/follow", friendshipHandler.Unfollow)
	authed.GET("/friendships/:userId/status", friendshipHandler.Status)

	// Phase 21: Notification
	notificationRepo := repository.NewNotificationRepository(db)
	notificationHandler := NewNotificationHandler(
		usecase.NewListNotificationsUseCase(notificationRepo),
		usecase.NewMarkNotificationReadUseCase(notificationRepo),
		usecase.NewMarkAllNotificationsReadUseCase(notificationRepo),
		usecase.NewCountUnreadNotificationsUseCase(notificationRepo),
	)
	authed.GET("/notifications", notificationHandler.List)
	authed.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	authed.PATCH("/notifications/:id/read", notificationHandler.MarkRead)
	authed.PUT("/notifications/:id/read", notificationHandler.MarkRead)
	authed.PATCH("/notifications/read-all", notificationHandler.MarkAllRead)
	authed.PUT("/notifications/read-all", notificationHandler.MarkAllRead)

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
	// /daily-goals/today はフロント MenuPage が呼ぶ。handler 側で path の :userId が "today" や数値以外なら
	// current user に解決する仕組みを入れている (see daily_goal_handler.go resolveUserID)。
	authed.GET("/daily-goals/:userId", dailyGoalHandler.Get)
	authed.PUT("/daily-goals/:userId", dailyGoalHandler.Upsert)

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

	// AiChat WebSocket (raw / native, Bedrock 連携は別 PR)
	// SockJS / STOMP は廃止し、フロントは ブラウザ標準 `WebSocket` で接続する。
	aiChatWsHandler := NewAiChatWsHandler()
	authed.GET("/ws/ai-chat", aiChatWsHandler.Handle)

	// Chat WebSocket (ルームごとブロードキャスト)。raw WebSocket + JSON プロトコル。
	chatWsHandler := NewChatWsHandler()
	authed.GET("/ws/chat/:roomId", chatWsHandler.Handle)

	return r
}
