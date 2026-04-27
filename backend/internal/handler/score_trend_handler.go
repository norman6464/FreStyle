package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ScoreTrendHandler struct {
	get *usecase.GetScoreTrendUseCase
}

func NewScoreTrendHandler(g *usecase.GetScoreTrendUseCase) *ScoreTrendHandler {
	return &ScoreTrendHandler{get: g}
}

func (h *ScoreTrendHandler) Get(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	trend, err := h.get.Execute(c.Request.Context(), uid, days)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}
