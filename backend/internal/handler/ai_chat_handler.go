package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AiChatHandler は AI チャット関連のエンドポイントを提供する。
type AiChatHandler struct {
	getSessions   *usecase.GetAiChatSessionsByUserIDUseCase
	createSession *usecase.CreateAiChatSessionUseCase
}

func NewAiChatHandler(
	getSessions *usecase.GetAiChatSessionsByUserIDUseCase,
	createSession *usecase.CreateAiChatSessionUseCase,
) *AiChatHandler {
	return &AiChatHandler{getSessions: getSessions, createSession: createSession}
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
