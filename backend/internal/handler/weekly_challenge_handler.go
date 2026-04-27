package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type WeeklyChallengeHandler struct {
	current  *usecase.GetCurrentWeeklyChallengeUseCase
	complete *usecase.CompleteWeeklyChallengeUseCase
}

func NewWeeklyChallengeHandler(c *usecase.GetCurrentWeeklyChallengeUseCase, comp *usecase.CompleteWeeklyChallengeUseCase) *WeeklyChallengeHandler {
	return &WeeklyChallengeHandler{current: c, complete: comp}
}

func (h *WeeklyChallengeHandler) GetCurrent(c *gin.Context) {
	got, err := h.current.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no_active_challenge"})
		return
	}
	c.JSON(http.StatusOK, got)
}

type completeChallengeReq struct {
	UserID      uint64 `json:"userId" binding:"required"`
	ChallengeID uint64 `json:"challengeId" binding:"required"`
}

func (h *WeeklyChallengeHandler) Complete(c *gin.Context) {
	var req completeChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.complete.Execute(c.Request.Context(), req.UserID, req.ChallengeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
