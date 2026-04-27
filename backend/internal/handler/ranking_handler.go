package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type RankingHandler struct {
	get *usecase.GetRankingUseCase
}

func NewRankingHandler(g *usecase.GetRankingUseCase) *RankingHandler {
	return &RankingHandler{get: g}
}

func (h *RankingHandler) Get(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	entries, err := h.get.Execute(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, entries)
}
