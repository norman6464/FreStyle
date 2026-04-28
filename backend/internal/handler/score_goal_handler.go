package handler

import (
	"errors"
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

var errInvalidUserID = errors.New("invalid userId")

// resolveUserID は path :userId が指定されていれば解析し、未指定なら current user を返す。
// 不正な :userId （非数値・0）は黙って current user にフォールバックさせず error を返す。
// これにより handler 側で 400 を返せる（IDOR 風の意図しない挙動を防ぐ）。
func (h *ScoreGoalHandler) resolveUserID(c *gin.Context) (uint64, error) {
	if v := c.Param("userId"); v != "" {
		uid, err := strconv.ParseUint(v, 10, 64)
		if err != nil || uid == 0 {
			return 0, errInvalidUserID
		}
		return uid, nil
	}
	return middleware.CurrentUserIDOrZero(c), nil
}

func (h *ScoreGoalHandler) Get(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
	uid, err := h.resolveUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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
