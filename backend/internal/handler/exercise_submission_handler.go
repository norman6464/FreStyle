package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// ExerciseSubmissionHandler は master_exercises に対する提出 API を扱う。
//
// API パス:
//
//	POST /api/v2/exercises/:slug/submit       コードを提出して採点
//	GET  /api/v2/exercises/:slug/submissions  current user の履歴
type ExerciseSubmissionHandler struct {
	submit *usecase.SubmitMasterExerciseUseCase
	list   *usecase.ListUserMasterSubmissionsUseCase
}

func NewExerciseSubmissionHandler(
	submit *usecase.SubmitMasterExerciseUseCase,
	list *usecase.ListUserMasterSubmissionsUseCase,
) *ExerciseSubmissionHandler {
	return &ExerciseSubmissionHandler{submit: submit, list: list}
}

type submitExerciseRequest struct {
	Code string `json:"code" binding:"required"`
}

// Submit は POST /api/v2/exercises/:slug/submit 。
func (h *ExerciseSubmissionHandler) Submit(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	var req submitExerciseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.submit.Execute(c.Request.Context(), usecase.SubmitMasterExerciseInput{
		UserID: uid,
		Slug:   slug,
		Code:   req.Code,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "コードの採点に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, out)
}

// List は GET /api/v2/exercises/:slug/submissions 。
func (h *ExerciseSubmissionHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), usecase.ListUserMasterSubmissionsInput{UserID: uid, Slug: slug})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "履歴の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
