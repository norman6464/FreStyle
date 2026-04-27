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
	// userId クエリ未指定 / parse 失敗時は空配列で返す（フロント未連携時の暫定対応）。
	// 本来は middleware で current user を解決して付与すべきだが、別 issue に分離。
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	if uid == 0 {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
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
