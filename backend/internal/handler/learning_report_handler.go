package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type LearningReportHandler struct {
	list    *usecase.ListLearningReportsUseCase
	request *usecase.RequestLearningReportUseCase
}

func NewLearningReportHandler(l *usecase.ListLearningReportsUseCase, r *usecase.RequestLearningReportUseCase) *LearningReportHandler {
	return &LearningReportHandler{list: l, request: r}
}

// List は current user の learning report 一覧を返す。
// userId はクライアントから受け取らない（IDOR 対策）。
func (h *LearningReportHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type requestReportReq struct {
	PeriodFrom string `json:"periodFrom" binding:"required"`
	PeriodTo   string `json:"periodTo" binding:"required"`
}

// Request は current user で report 生成を要求する。
func (h *LearningReportHandler) Request(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req requestReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	from, err := time.Parse(time.RFC3339, req.PeriodFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "periodFrom must be RFC3339"})
		return
	}
	to, err := time.Parse(time.RFC3339, req.PeriodTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "periodTo must be RFC3339"})
		return
	}
	got, err := h.request.Execute(c.Request.Context(), usecase.RequestLearningReportInput{
		UserID: uid, PeriodFrom: from, PeriodTo: to,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, got)
}
