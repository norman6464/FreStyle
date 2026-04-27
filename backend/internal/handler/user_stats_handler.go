package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type UserStatsHandler struct {
	get *usecase.GetUserStatsUseCase
}

func NewUserStatsHandler(g *usecase.GetUserStatsUseCase) *UserStatsHandler {
	return &UserStatsHandler{get: g}
}

func (h *UserStatsHandler) Get(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	stats, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
