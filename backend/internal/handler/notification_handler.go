package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NotificationHandler struct {
	list     *usecase.ListNotificationsUseCase
	markRead *usecase.MarkNotificationReadUseCase
}

func NewNotificationHandler(l *usecase.ListNotificationsUseCase, m *usecase.MarkNotificationReadUseCase) *NotificationHandler {
	return &NotificationHandler{list: l, markRead: m}
}

func (h *NotificationHandler) List(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.markRead.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
