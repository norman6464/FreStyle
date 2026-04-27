package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type SessionNoteHandler struct {
	get    *usecase.GetSessionNoteUseCase
	upsert *usecase.UpsertSessionNoteUseCase
}

func NewSessionNoteHandler(g *usecase.GetSessionNoteUseCase, u *usecase.UpsertSessionNoteUseCase) *SessionNoteHandler {
	return &SessionNoteHandler{get: g, upsert: u}
}

func (h *SessionNoteHandler) Get(c *gin.Context) {
	sid, _ := strconv.ParseUint(c.Param("sessionId"), 10, 64)
	n, err := h.get.Execute(c.Request.Context(), sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if n == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

type sessionNoteUpsertReq struct {
	UserID  uint64 `json:"userId" binding:"required"`
	Content string `json:"content"`
}

func (h *SessionNoteHandler) Upsert(c *gin.Context) {
	sid, _ := strconv.ParseUint(c.Param("sessionId"), 10, 64)
	var req sessionNoteUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), usecase.UpsertSessionNoteInput{
		SessionID: sid, UserID: req.UserID, Content: req.Content,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
