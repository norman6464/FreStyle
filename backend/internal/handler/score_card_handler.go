package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ScoreCardHandler struct {
	list   *usecase.ListScoreCardsByUserIDUseCase
	create *usecase.CreateScoreCardUseCase
}

func NewScoreCardHandler(l *usecase.ListScoreCardsByUserIDUseCase, c *usecase.CreateScoreCardUseCase) *ScoreCardHandler {
	return &ScoreCardHandler{list: l, create: c}
}

func (h *ScoreCardHandler) List(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *ScoreCardHandler) Create(c *gin.Context) {
	var card domain.ScoreCard
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), &card)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}
