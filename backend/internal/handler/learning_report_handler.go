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

// @Summary      学習 レポート 一覧
// @Description  current user の レポート を 期間 降順 で 返す。 userId は IDOR 対策 で 受け取らない。
// @Tags         learning-reports
// @Produce      json
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.LearningReport
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /learning-reports [get]
// @Security     CookieAuth
func (h *LearningReportHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// requestReportReq は月次レポート生成 body。year + month から月初〜翌月初の period を組み立てる。
type requestReportReq struct {
	Year  int `json:"year"  binding:"required,min=2000,max=2100"`
	Month int `json:"month" binding:"required,min=1,max=12"`
}

// @Summary      月次 学習 レポート 生成 要求
// @Description  current user で 指定 月 の レポート 生成 ジョブ を 受け付け、 SQS に enqueue (現状 stub)。 202 Accepted を 返す。
// @Tags         learning-reports
// @Accept       json
// @Produce      json
// @Param        body  body      requestReportReq  true  "year + month"
// @Success      202   {object}  github_com_norman6464_FreStyle_backend_internal_domain.LearningReport
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Router       /learning-reports/generate [post]
// @Security     CookieAuth
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
	// year + month を月初〜翌月初の半開区間 [PeriodFrom, PeriodTo) に変換する。
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
