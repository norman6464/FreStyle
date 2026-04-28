package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NotificationHandler struct {
	list     *usecase.ListNotificationsUseCase
	markRead *usecase.MarkNotificationReadUseCase
}

func NewNotificationHandler(l *usecase.ListNotificationsUseCase, m *usecase.MarkNotificationReadUseCase) *NotificationHandler {
	return &NotificationHandler{list: l, markRead: m}
}

// List は常に認証済 current user の通知を返す。
// 過去は ?userId= クエリを受け付けていたが、authz 機構が無いため任意の userId を
// クエリで指定できると IDOR になる。admin 機構が入るまでは current user 固定にする。
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

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.markRead.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
