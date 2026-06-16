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
//
//	@Summary      自分の学習進捗（完了レッスン一覧）
//	@Description  current user が完了した教材（レッスン）の一覧を返す。進捗バー / 完了チェック表示用。userId は受け取らない（current user 固定）。
//	@Tags         lesson-progress
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.UserLessonProgress
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /lesson-progress [get]
//	@Security     CookieAuth
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
//
//	@Summary      レッスンを完了にする
//	@Description  current user 名義で教材（レッスン）を完了として記録する。冪等（二重実行しても 1 件）。course は教材から解決する。
//	@Tags         lesson-progress
//	@Accept       json
//	@Produce      json
//	@Param        body  body      markLessonCompleteRequest  true  "完了する教材 ID"
//	@Success      204   "成功（本文なし）"
//	@Failure      400   {object}  errorResponse  "不正な body"
//	@Failure      401   {object}  errorResponse  "未認証"
//	@Failure      404   {object}  errorResponse  "教材が存在しない"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /lesson-progress [post]
//	@Security     CookieAuth
func (h *LessonProgressHandler) Complete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req markLessonCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	err := h.complete.Execute(c.Request.Context(), uid, req.TeachingMaterialID)
	switch {
	case errors.Is(err, usecase.ErrLessonNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "lesson_not_found"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// Incomplete は完了記録を取り消す。
//
//	@Summary      レッスンの完了を取り消す
//	@Description  current user の当該教材の完了記録を取り消す（未記録でも 204）。
//	@Tags         lesson-progress
//	@Produce      json
//	@Param        teachingMaterialId  path  int  true  "教材 ID"
//	@Success      204   "成功（本文なし）"
//	@Failure      400   {object}  errorResponse  "不正な ID"
//	@Failure      401   {object}  errorResponse  "未認証"
//	@Failure      500   {object}  errorResponse  "DB 失敗"
//	@Router       /lesson-progress/{teachingMaterialId} [delete]
//	@Security     CookieAuth
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
