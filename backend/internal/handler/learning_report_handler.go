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

// requestReportReq は POST /learning-reports/generate のリクエスト body。
//
// フロントエンドは月次サマリの生成しか叩かないため、 期間指定は year + month の
// 1 形式に揃える。 バックエンド側で月初〜翌月初の period を組み立てて usecase に渡す。
type requestReportReq struct {
	Year  int `json:"year"  binding:"required,min=2000,max=2100"`
	Month int `json:"month" binding:"required,min=1,max=12"`
}

// Request は current user で月次 learning report 生成を要求する。
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
	// year + month → 月初 [00:00:00 UTC] 〜 翌月初 [00:00:00 UTC] の期間に変換。
	// レポート集計の実装が PeriodFrom <= submitted_at < PeriodTo の半開区間で扱う前提。
	from := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0)
	got, err := h.request.Execute(c.Request.Context(), usecase.RequestLearningReportInput{
		UserID: uid, PeriodFrom: from, PeriodTo: to,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, got)
}
