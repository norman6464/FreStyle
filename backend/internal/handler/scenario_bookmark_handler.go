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
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	if uid == 0 {
		uid = middleware.MustCurrentUserID(c)
	}
	if uid == 0 {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type addBookmarkReq struct {
	UserID     uint64 `json:"userId" binding:"required"`
	ScenarioID uint64 `json:"scenarioId" binding:"required"`
}

func (h *ScenarioBookmarkHandler) Add(c *gin.Context) {
	var req addBookmarkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.add.Execute(c.Request.Context(), req.UserID, req.ScenarioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *ScenarioBookmarkHandler) Remove(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	sid, _ := strconv.ParseUint(c.Param("scenarioId"), 10, 64)
	if err := h.remove.Execute(c.Request.Context(), uid, sid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
