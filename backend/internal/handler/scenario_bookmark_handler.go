package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ScenarioBookmarkHandler struct {
	list   *usecase.ListScenarioBookmarksUseCase
	add    *usecase.AddScenarioBookmarkUseCase
	remove *usecase.RemoveScenarioBookmarkUseCase
}

func NewScenarioBookmarkHandler(
	l *usecase.ListScenarioBookmarksUseCase,
	a *usecase.AddScenarioBookmarkUseCase,
	r *usecase.RemoveScenarioBookmarkUseCase,
) *ScenarioBookmarkHandler {
	return &ScenarioBookmarkHandler{list: l, add: a, remove: r}
}

func (h *ScenarioBookmarkHandler) List(c *gin.Context) {
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

// Add は path :scenarioId を current user の bookmark として登録する。
// userId はクライアントから受け取らずサーバ側で current user に固定する（IDOR 対策）。
func (h *ScenarioBookmarkHandler) Add(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sid, err := strconv.ParseUint(c.Param("scenarioId"), 10, 64)
	if err != nil || sid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenarioId"})
		return
	}
	got, err := h.add.Execute(c.Request.Context(), uid, sid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

// Remove は path :scenarioId を current user の bookmark として削除する。
// 他人の bookmark を消せないように DB の WHERE 句で user_id を絞る。
func (h *ScenarioBookmarkHandler) Remove(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	sid, err := strconv.ParseUint(c.Param("scenarioId"), 10, 64)
	if err != nil || sid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid scenarioId"})
		return
	}
	if err := h.remove.Execute(c.Request.Context(), uid, sid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
