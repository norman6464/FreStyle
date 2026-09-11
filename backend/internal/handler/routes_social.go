package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/usecase/notification"
)

// registerSocialRoutes は通知の REST エンドポイントを登録する。
// Friendship / フォロー機能は削除済み。
func registerSocialRoutes(g *gin.RouterGroup, deps *routeDeps) {
	notificationRepo := persistence.NewNotificationRepository(deps.db)
	notificationHandler := NewNotificationHandler(
		notification.NewListNotificationsUseCase(notificationRepo),
		notification.NewMarkNotificationReadUseCase(notificationRepo),
		notification.NewMarkAllNotificationsReadUseCase(notificationRepo),
		notification.NewCountUnreadNotificationsUseCase(notificationRepo),
	)
	g.GET("/notifications", notificationHandler.List)
	g.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	g.PATCH("/notifications/:id/read", notificationHandler.MarkRead)
	g.PUT("/notifications/:id/read", notificationHandler.MarkRead) //apispec:allow フロント互換の別 method（正規は PATCH）
	g.PATCH("/notifications/read-all", notificationHandler.MarkAllRead)
	g.PUT("/notifications/read-all", notificationHandler.MarkAllRead) //apispec:allow フロント互換の別 method（正規は PATCH）
}
