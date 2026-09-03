package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// LessonProgressHandler は trainee 自身の教材（レッスン）完了状態を扱う。
// すべて current user 名義で、他人の進捗は操作・閲覧できない（userId は受け取らない）。
type LessonProgressHandler struct {
	complete   *usecase.MarkLessonCompletedUseCase
	incomplete *usecase.MarkLessonIncompleteUseCase
	list       *usecase.ListLessonProgressUseCase
}

func NewLessonProgressHandler(
	c *usecase.MarkLessonCompletedUseCase,
	i *usecase.MarkLessonIncompleteUseCase,
	l *usecase.ListLessonProgressUseCase,
) *LessonProgressHandler {
	return &LessonProgressHandler{complete: c, incomplete: i, list: l}
}

type markLessonCompleteRequest struct {
	TeachingMaterialID uint64 `json:"teachingMaterialId" binding:"required"`
}

// List は current user の完了済みレッスン一覧を返す。
func (h *LessonProgressHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// Complete は教材を完了として記録する（冪等）。
func (h *LessonProgressHandler) Complete(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req markLessonCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	err := h.complete.Execute(c.Request.Context(), usecase.MarkLessonCompletedInput{
		UserID:             user.ID,
		ActorWorkspace:     user.WorkspaceRef(),
		ActorRole:          user.Role,
		TeachingMaterialID: req.TeachingMaterialID,
	})
	switch {
	case errors.Is(err, usecase.ErrLessonNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "lesson_not_found"})
		return
	case errors.Is(err, usecase.ErrLessonForbidden):
		c.JSON(http.StatusForbidden, errorResponse{Error: "lesson_forbidden"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Incomplete は完了記録を取り消す。
func (h *LessonProgressHandler) Incomplete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	mid, err := strconv.ParseUint(c.Param("teachingMaterialId"), 10, 64)
	if err != nil || mid == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}
	if err := h.incomplete.Execute(c.Request.Context(), uid, mid); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}
