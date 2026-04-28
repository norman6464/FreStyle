package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type DailyGoalHandler struct {
	get    *usecase.GetDailyGoalUseCase
	upsert *usecase.UpsertDailyGoalUseCase
}

func NewDailyGoalHandler(g *usecase.GetDailyGoalUseCase, u *usecase.UpsertDailyGoalUseCase) *DailyGoalHandler {
	return &DailyGoalHandler{get: g, upsert: u}
}

// resolveUserID は path の :userId が "today" や空文字なら current user を返す。
// フロントが /api/v2/daily-goals/today を叩くケースに対応。
func (h *DailyGoalHandler) resolveUserID(c *gin.Context) uint64 {
	param := c.Param("userId")
	if uid, err := strconv.ParseUint(param, 10, 64); err == nil && uid > 0 {
		return uid
	}
	return middleware.CurrentUserIDOrZero(c)
}

func (h *DailyGoalHandler) Get(c *gin.Context) {
	uid := h.resolveUserID(c)
	if uid == 0 {
		c.JSON(http.StatusOK, gin.H{"date": "", "targetMinutes": 0, "actualMinutes": 0, "isAchieved": false})
		return
	}
	dateStr := c.DefaultQuery("date", time.Now().UTC().Format("2006-01-02"))
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
		return
	}
	got, err := h.get.Execute(c.Request.Context(), uid, date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if got == nil {
		c.JSON(http.StatusOK, gin.H{"userId": uid, "date": dateStr, "targetMinutes": 0, "actualMinutes": 0, "isAchieved": false})
		return
	}
	c.JSON(http.StatusOK, got)
}

type dailyGoalReq struct {
	Date          string `json:"date" binding:"required"`
	TargetMinutes int    `json:"targetMinutes"`
	ActualMinutes int    `json:"actualMinutes"`
}

func (h *DailyGoalHandler) Upsert(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	var req dailyGoalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be YYYY-MM-DD"})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), usecase.UpsertDailyGoalInput{
		UserID: uid, Date: date, TargetMin: req.TargetMinutes, ActualMin: req.ActualMinutes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
