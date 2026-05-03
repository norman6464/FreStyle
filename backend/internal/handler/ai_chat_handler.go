package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// AiChatHandler は AI チャット関連のエンドポイントを提供する。
type AiChatHandler struct {
	getSessions   *usecase.GetAiChatSessionsByUserIDUseCase
	createSession *usecase.CreateAiChatSessionUseCase
	getSession    *usecase.GetAiChatSessionUseCase
	updateTitle   *usecase.UpdateAiChatSessionTitleUseCase
	deleteSession *usecase.DeleteAiChatSessionUseCase
	getMessages   *usecase.GetAiChatMessagesUseCase
}

func NewAiChatHandler(
	getSessions *usecase.GetAiChatSessionsByUserIDUseCase,
	createSession *usecase.CreateAiChatSessionUseCase,
	getSession *usecase.GetAiChatSessionUseCase,
	updateTitle *usecase.UpdateAiChatSessionTitleUseCase,
	deleteSession *usecase.DeleteAiChatSessionUseCase,
	getMessages *usecase.GetAiChatMessagesUseCase,
) *AiChatHandler {
	return &AiChatHandler{
		getSessions:   getSessions,
		createSession: createSession,
		getSession:    getSession,
		updateTitle:   updateTitle,
		deleteSession: deleteSession,
		getMessages:   getMessages,
	}
}

// GetSessions は GET /ai-chat/sessions
// userId はクライアントから受け取らず、middleware の current user を使う。
func (h *AiChatHandler) GetSessions(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.getSessions.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createSessionReq struct {
	Title       string  `json:"title" binding:"required"`
	SessionType string  `json:"sessionType"`
	ScenarioID  *uint64 `json:"scenarioId"`
}

// CreateSession は POST /ai-chat/sessions
// userId はクライアントから受け取らず current user で固定する（IDOR 対策）。
func (h *AiChatHandler) CreateSession(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req createSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.createSession.Execute(c.Request.Context(), usecase.CreateAiChatSessionInput{
		UserID:      uid,
		Title:       req.Title,
		SessionType: req.SessionType,
		ScenarioID:  req.ScenarioID,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

// parseSessionID は :id パスパラメータを uint64 に変換する共通ヘルパー。
func parseSessionID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return 0, false
	}
	return id, true
}

// GetSession は GET /ai-chat/sessions/:id
func (h *AiChatHandler) GetSession(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	s, err := h.getSession.Execute(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, s)
}

type updateSessionTitleReq struct {
	Title string `json:"title" binding:"required"`
}

// UpdateSessionTitle は PUT /ai-chat/sessions/:id
func (h *AiChatHandler) UpdateSessionTitle(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	var req updateSessionTitleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.updateTitle.Execute(c.Request.Context(), id, req.Title); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	s, _ := h.getSession.Execute(c.Request.Context(), id)
	c.JSON(http.StatusOK, s)
}

// DeleteSession は DELETE /ai-chat/sessions/:id
func (h *AiChatHandler) DeleteSession(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	if err := h.deleteSession.Execute(c.Request.Context(), id, uid); err != nil {
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetMessages は GET /ai-chat/sessions/:id/messages
func (h *AiChatHandler) GetMessages(c *gin.Context) {
	id, ok := parseSessionID(c)
	if !ok {
		return
	}
	msgs, err := h.getMessages.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, msgs)
}
