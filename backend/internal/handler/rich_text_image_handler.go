package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/richtextimage"
)

type RichTextImageHandler struct {
	issue *richtextimage.IssueRichTextImageUploadURLUseCase
}

func NewRichTextImageHandler(i *richtextimage.IssueRichTextImageUploadURLUseCase) *RichTextImageHandler {
	return &RichTextImageHandler{issue: i}
}

// issueUploadURLReq は body 受け取り。userId は受け取らず middleware の current user を使う（IDOR 対策）。
type issueUploadURLReq struct {
	ContentType string `json:"contentType"`
	// Size はバイト数（FRESTYLE-9: サイズ上限の検証に使う）。省略時は 0 になり、
	// domain.ValidateImageUpload が「0 以下は拒否」で弾く。
	Size int64 `json:"size"`
}

func (h *RichTextImageHandler) IssueUploadURL(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req issueUploadURLReq
	// body 無し (EOF) は許容するが、不正 JSON は 400 で弾く。
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.issue.Execute(c.Request.Context(), uid, req.ContentType, req.Size)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUnsupportedImageContentType):
			c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_content_type"})
		case errors.Is(err, domain.ErrImageTooLarge):
			c.JSON(http.StatusBadRequest, gin.H{"error": "image_too_large"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, got)
}
