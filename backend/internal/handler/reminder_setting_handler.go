package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ReminderSettingHandler struct {
	get    *usecase.GetReminderSettingUseCase
	upsert *usecase.UpsertReminderSettingUseCase
}

func NewReminderSettingHandler(g *usecase.GetReminderSettingUseCase, u *usecase.UpsertReminderSettingUseCase) *ReminderSettingHandler {
	return &ReminderSettingHandler{get: g, upsert: u}
}

func (h *ReminderSettingHandler) Get(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	s, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, s)
}

type reminderUpsertReq struct {
	HourLocal    int  `json:"hourLocal"`
	MinuteLocal  int  `json:"minuteLocal"`
	WeekdaysMask int  `json:"weekdaysMask"`
	IsActive     bool `json:"isActive"`
}

func (h *ReminderSettingHandler) Upsert(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	var req reminderUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), usecase.UpsertReminderSettingInput{
		UserID: uid, HourLocal: req.HourLocal, MinuteLocal: req.MinuteLocal,
		WeekdaysMask: req.WeekdaysMask, IsActive: req.IsActive,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
