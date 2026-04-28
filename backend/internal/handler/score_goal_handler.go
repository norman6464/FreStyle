package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ScoreGoalHandler struct {
	get    *usecase.GetScoreGoalUseCase
	upsert *usecase.UpsertScoreGoalUseCase
}

func NewScoreGoalHandler(g *usecase.GetScoreGoalUseCase, u *usecase.UpsertScoreGoalUseCase) *ScoreGoalHandler {
	return &ScoreGoalHandler{get: g, upsert: u}
}

// resolveUserID は path :userId か、未指定なら current user の users.id を返す。
func (h *ScoreGoalHandler) resolveUserID(c *gin.Context) uint64 {
	if v := c.Param("userId"); v != "" {
		if uid, err := strconv.ParseUint(v, 10, 64); err == nil {
			return uid
		}
	}
	return middleware.MustCurrentUserID(c)
}

func (h *ScoreGoalHandler) Get(c *gin.Context) {
	uid := h.resolveUserID(c)
	if uid == 0 {
		c.JSON(http.StatusOK, gin.H{"targetScore": 0})
		return
	}
	g, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if g == nil {
		// 未設定でも 0 で返す（フロントが落ちないようにする）。
		c.JSON(http.StatusOK, gin.H{"userId": uid, "targetScore": 0})
		return
	}
	c.JSON(http.StatusOK, g)
}

type scoreGoalUpsertReq struct {
	TargetScore float64 `json:"targetScore"`
}

func (h *ScoreGoalHandler) Upsert(c *gin.Context) {
	uid := h.resolveUserID(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req scoreGoalUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.upsert.Execute(c.Request.Context(), uid, req.TargetScore)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
