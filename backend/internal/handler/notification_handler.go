package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NotificationHandler struct {
	list        *usecase.ListNotificationsUseCase
	markRead    *usecase.MarkNotificationReadUseCase
	markAllRead *usecase.MarkAllNotificationsReadUseCase
	countUnread *usecase.CountUnreadNotificationsUseCase
}

func NewNotificationHandler(
	l *usecase.ListNotificationsUseCase,
	m *usecase.MarkNotificationReadUseCase,
	a *usecase.MarkAllNotificationsReadUseCase,
	cu *usecase.CountUnreadNotificationsUseCase,
) *NotificationHandler {
	return &NotificationHandler{list: l, markRead: m, markAllRead: a, countUnread: cu}
}

// List は常に認証済 current user の通知を返す。
func (h *NotificationHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// MarkRead は所有者検証つきで通知を既読化する。
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.markRead.Execute(c.Request.Context(), uid, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// MarkAllRead は current user の全通知をまとめて既読化する。
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if err := h.markAllRead.Execute(c.Request.Context(), uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// UnreadCount は current user の未読通知数を整数で返す。
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	n, err := h.countUnread.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, n)
}
