package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AiChatHandler は AI チャット関連のエンドポイントを提供する。
// Spring Boot の AiChatController に相当。
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
func (h *AiChatHandler) GetSessions(c *gin.Context) {
	userIDStr := c.Query("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}
	rows, err := h.getSessions.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createSessionReq struct {
	UserID      uint64  `json:"userId"`
	Title       string  `json:"title" binding:"required"`
	SessionType string  `json:"sessionType"`
	ScenarioID  *uint64 `json:"scenarioId"`
}

// CreateSession は POST /ai-chat/sessions
func (h *AiChatHandler) CreateSession(c *gin.Context) {
	var req createSessionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.createSession.Execute(c.Request.Context(), usecase.CreateAiChatSessionInput{
		UserID:      req.UserID,
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
