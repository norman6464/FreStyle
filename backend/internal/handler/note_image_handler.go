package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NoteImageHandler struct {
	issue *usecase.IssueNoteImageUploadURLUseCase
}

func NewNoteImageHandler(i *usecase.IssueNoteImageUploadURLUseCase) *NoteImageHandler {
	return &NoteImageHandler{issue: i}
}

// issueUploadURLReq は body 受け取り。 userId は **受け取らない** (current user
// を middleware 経由 で 強制)。 OpenAPI Phase 3 で IDOR を 修正。
type issueUploadURLReq struct {
	ContentType string `json:"contentType"`
}

// @Summary      ノート 画像 PUT 署名 URL
// @Description  current user 用 の S3 PUT 署名 URL を 発行。 userId は body から 受け取らず middleware の current user を 使う (IDOR 対策、 Phase 3 で 修正)。
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        body  body      issueUploadURLReq  false  "contentType (任意)"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.NoteImageUploadURL
// @Failure      400   {object}  errorResponse  "発行 失敗"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Router       /notes/image-upload-url [post]
// @Security     CookieAuth
func (h *NoteImageHandler) IssueUploadURL(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req issueUploadURLReq
	// body 無し (EOF) は 許容 する が、 不正 JSON / 型 違い は 400 で 弾く
	// (silently masking malformed JSON を 避ける)。
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.issue.Execute(c.Request.Context(), uid, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
