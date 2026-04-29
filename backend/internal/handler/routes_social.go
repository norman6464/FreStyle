package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerSocialRoutes は Friendship + 単方向フォローと通知の REST エンドポイントを登録する。
func registerSocialRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 20: Friendship + 単方向フォロー（フロントの follow / unfollow / status / following / followers に対応）
	friendshipRepo := repository.NewFriendshipRepository(deps.db)
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
	g.GET("/friendships", friendshipHandler.List)
	g.POST("/friendships", friendshipHandler.Request)
	g.PATCH("/friendships/:id", friendshipHandler.Respond)
	g.GET("/friendships/following", friendshipHandler.Following)
	g.GET("/friendships/followers", friendshipHandler.Followers)
	g.POST("/friendships/:userId/follow", friendshipHandler.Follow)
	g.DELETE("/friendships/:userId/follow", friendshipHandler.Unfollow)
	g.GET("/friendships/:userId/status", friendshipHandler.Status)

	// Phase 21: Notification
	notificationRepo := repository.NewNotificationRepository(deps.db)
	notificationHandler := NewNotificationHandler(
		usecase.NewListNotificationsUseCase(notificationRepo),
		usecase.NewMarkNotificationReadUseCase(notificationRepo),
		usecase.NewMarkAllNotificationsReadUseCase(notificationRepo),
		usecase.NewCountUnreadNotificationsUseCase(notificationRepo),
	)
	g.GET("/notifications", notificationHandler.List)
	g.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	g.PATCH("/notifications/:id/read", notificationHandler.MarkRead)
	g.PUT("/notifications/:id/read", notificationHandler.MarkRead)
	g.PATCH("/notifications/read-all", notificationHandler.MarkAllRead)
	g.PUT("/notifications/read-all", notificationHandler.MarkAllRead)
}
