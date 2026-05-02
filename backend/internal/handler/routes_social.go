package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerSocialRoutes は通知の REST エンドポイントを登録する。
// Friendship / フォロー機能は削除済み。
func registerSocialRoutes(g *gin.RouterGroup, deps *routeDeps) {
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
