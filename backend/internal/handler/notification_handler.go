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
//
//	@Summary      通知 一覧
//	@Description  current user の 通知 を 作成 日 降順 で 返す。
//	@Tags         notifications
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.Notification
//	@Failure      400  {object}  errorResponse  "DB 取得 失敗"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Router       /notifications [get]
//	@Security     CookieAuth
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
//
//	@Summary      通知 単一 既読 化
//	@Description  指定 通知 を 既読 に する (所有者 検証 込み)。 PATCH / PUT 両方 受け付ける。
//	@Tags         notifications
//	@Param        id  path  int  true  "通知 ID"
//	@Success      204  "成功 (本文 なし)"
//	@Failure      400  {object}  errorResponse  "DB 失敗"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Router       /notifications/{id}/read [patch]
//	@Security     CookieAuth
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
//
//	@Summary      通知 一括 既読 化
//	@Description  current user の 全 未読 通知 を まとめて 既読 に する。 PATCH / PUT 両方 受け付ける。
//	@Tags         notifications
//	@Success      204  "成功 (本文 なし)"
//	@Failure      400  {object}  errorResponse  "DB 失敗"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Router       /notifications/read-all [patch]
//	@Security     CookieAuth
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
//
//	@Summary      未読 通知 数
//	@Description  current user の 未読 通知 数 を 整数 で 返す (バッジ 表示 用)。
//	@Tags         notifications
//	@Produce      json
//	@Success      200  {integer}  int64
//	@Failure      400  {object}   errorResponse  "DB 失敗"
//	@Failure      401  {object}   errorResponse  "未 認証"
//	@Router       /notifications/unread-count [get]
//	@Security     CookieAuth
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
