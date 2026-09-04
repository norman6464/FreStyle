package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// ChapterViewHandler はユーザーの章（教材）閲覧記録 API を扱う。
type ChapterViewHandler struct {
	record *usecase.RecordChapterViewUseCase
}

func NewChapterViewHandler(r *usecase.RecordChapterViewUseCase) *ChapterViewHandler {
	return &ChapterViewHandler{record: r}
}

// RecordView は章を開いた（閲覧した）ことを記録する。
// フロントエンドは教材ページを開いた瞬間にこのエンドポイントを叩く。
// 失敗してもユーザー体験に影響しないため、エラーは 204 で透過させる。
func (h *ChapterViewHandler) RecordView(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	mid, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || mid == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}
	// ベストエフォート — 失敗しても 204 で返す。
	_ = h.record.Execute(c.Request.Context(), usecase.RecordChapterViewInput{
		UserID:             user.ID,
		ActorWorkspace:     user.WorkspaceRef(),
		TeachingMaterialID: mid,
	})
	c.Status(http.StatusNoContent)
}
