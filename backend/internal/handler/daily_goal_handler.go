package handler

import (
	"errors"
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

var (
	errDailyGoalForbidden    = errors.New("forbidden")
	errDailyGoalUnauthorized = errors.New("unauthorized")
)

// resolveUserID は path :userId を current user と突き合わせる。
//   - "today" / 空文字 / 数値以外 → current user (フロントの /daily-goals/today に対応)
//   - 数値 0 → 不正
//   - 数値が current user 以外 → forbidden（IDOR 対策）
//   - 数値が current user と一致 → そのまま返す
func (h *DailyGoalHandler) resolveUserID(c *gin.Context) (uint64, error) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		return 0, errDailyGoalUnauthorized
	}
	param := c.Param("userId")
	if param == "" || param == "today" {
		return cur, nil
	}
	uid, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		// 数字でない path リテラル（例 "today" 以外の文字列）も current user として扱う。
		return cur, nil
	}
	if uid == 0 {
		return 0, errDailyGoalForbidden
	}
	if uid != cur {
		return 0, errDailyGoalForbidden
	}
	return uid, nil
}

func (h *DailyGoalHandler) Get(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeDailyGoalError(c, err)
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
		// 未設定でも success と同形 (userId / date / targetMinutes / actualMinutes / isAchieved) を返す。
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
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeDailyGoalError(c, err)
		return
	}
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

func writeDailyGoalError(c *gin.Context, err error) {
	switch err {
	case errDailyGoalUnauthorized:
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errDailyGoalForbidden:
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
