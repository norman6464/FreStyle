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

// issueUploadURLReq は body 受け取り。userId は受け取らず middleware の current user を使う（IDOR 対策）。
type issueUploadURLReq struct {
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// @Summary      ノート 画像 PUT 署名 URL
// @Description  current user 用 の S3 PUT 署名 URL を 発行。 userId は body から 受け取らず middleware の current user を 使う (IDOR 対策、 Phase 3 で 修正)。 contentType は 画像 MIME (png/jpeg/gif/webp) のみ、 sizeBytes は 上限 5MB を 事前 検証。
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        body  body      issueUploadURLReq  true  "contentType / sizeBytes"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.NoteImageUploadURL
// @Failure      400   {object}  errorResponse  "発行 失敗"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      413   {object}  errorResponse  "サイズ 上限 超過"
// @Failure      415   {object}  errorResponse  "未 サポート MIME"
// @Router       /notes/images/upload-url [post]
// @Security     CookieAuth
func (h *NoteImageHandler) IssueUploadURL(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req issueUploadURLReq
	// body 無し (EOF) は下の必須チェックで 400 にするが、不正 JSON はここで弾く。
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ContentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contentType is required"})
		return
	}
	if req.SizeBytes <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sizeBytes must be positive"})
		return
	}
	got, err := h.issue.Execute(c.Request.Context(), usecase.IssueNoteImageUploadURLInput{
		UserID:      uid,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrNoteImageUnsupportedType):
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrNoteImageTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, got)
}
