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
//
//	@Summary      章の閲覧を記録する
//	@Description  ユーザーが教材（章）を開いたときに呼び出す。user_chapter_views を upsert し「続きから」カードの基盤データを更新する。 エラーは握り潰し 204 を返す（ベストエフォート）。
//	@Tags         teaching-materials
//	@Produce      json
//	@Param        id  path  int  true  "教材 ID"
//	@Success      204  "記録成功（本文なし）"
//	@Failure      400  {object}  errorResponse  "不正な ID"
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Router       /teaching-materials/{id}/view [post]
//	@Security     CookieAuth
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
	companyID := user.CompanyIDValue()
	// ベストエフォート — 失敗しても 204 で返す。
	_ = h.record.Execute(c.Request.Context(), usecase.RecordChapterViewInput{
		UserID:             user.ID,
		ActorCompanyID:     companyID,
		ActorRole:          user.Role,
		TeachingMaterialID: mid,
	})
	c.Status(http.StatusNoContent)
}
