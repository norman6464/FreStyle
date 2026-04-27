package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ScoreGoalHandler struct {
	get    *usecase.GetScoreGoalUseCase
	upsert *usecase.UpsertScoreGoalUseCase
}

func NewScoreGoalHandler(g *usecase.GetScoreGoalUseCase, u *usecase.UpsertScoreGoalUseCase) *ScoreGoalHandler {
	return &ScoreGoalHandler{get: g, upsert: u}
}

func (h *ScoreGoalHandler) Get(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	g, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

type scoreGoalUpsertReq struct {
	TargetScore float64 `json:"targetScore"`
}

func (h *ScoreGoalHandler) Upsert(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	var req scoreGoalUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), uid, req.TargetScore)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
