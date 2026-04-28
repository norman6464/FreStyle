package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
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
	// userId は (1) クエリ → (2) middleware の current user の順で解決する。
	// フロントは /api/v2/score-cards だけ叩けば良く、認証済 user の score-card 一覧が返る。
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	if uid == 0 {
		uid = middleware.CurrentUserIDOrZero(c)
	}
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
