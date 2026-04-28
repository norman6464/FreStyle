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

// List は ?userId= が指定されていても current user と一致しなければ 403。
// フロントは /api/v2/score-cards だけ叩けば current user の一覧が返る。
func (h *ScoreCardHandler) List(c *gin.Context) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if q := c.Query("userId"); q != "" {
		qUID, _ := strconv.ParseUint(q, 10, 64)
		if qUID != 0 && qUID != cur {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
	}
	rows, err := h.list.Execute(c.Request.Context(), cur)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// Create はクライアント値の card.UserID を信用せず current user で必ず上書きする（IDOR 対策）。
// クライアントが他 userId を指定して投げてきても、保存される行は必ず current user 名義になる。
func (h *ScoreCardHandler) Create(c *gin.Context) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var card domain.ScoreCard
	if err := c.ShouldBindJSON(&card); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	card.UserID = cur
	got, err := h.create.Execute(c.Request.Context(), &card)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}
