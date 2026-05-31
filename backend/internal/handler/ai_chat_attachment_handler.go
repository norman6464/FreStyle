package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AiChatAttachmentHandler は AI チャット添付の S3 PUT presigned URL を発行する。
type AiChatAttachmentHandler struct {
	issueUseCase *usecase.IssueAiChatAttachmentUploadURLUseCase
}

func NewAiChatAttachmentHandler(uc *usecase.IssueAiChatAttachmentUploadURLUseCase) *AiChatAttachmentHandler {
	return &AiChatAttachmentHandler{issueUseCase: uc}
}

type aiChatAttachmentUploadURLRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
}

// IssueUploadURL は添付の PUT presigned URL を発行する。
// 4xx は 400（バリデーション）/ 401 / 413（サイズ超過）/ 415（未対応 MIME）。
//
//	@Summary      AI チャット 添付 PUT 署名 URL
//	@Description  ai-chat/{userId}/{uuid}.{ext} の キー で S3 PUT 用 presigned URL を 発行。 contentType は image/png 等 の 許容 セット のみ。 sizeBytes 上限 (image 5MB / document 4.5MB) も 事前 検証。
//	@Tags         ai-chat
//	@Accept       json
//	@Produce      json
//	@Param        body  body      aiChatAttachmentUploadURLRequest  true  "filename / contentType / sizeBytes"
//	@Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_usecase_repository.AiChatAttachmentUploadURL
//	@Failure      400   {object}  errorResponse  "バリデーション 失敗"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      413   {object}  errorResponse  "サイズ 上限 超過"
//	@Failure      415   {object}  errorResponse  "未 サポート MIME"
//	@Failure      503   {object}  errorResponse  "添付 アップロード 機能 が 設定 されて いない (dev/stub)"
//	@Failure      500   {object}  errorResponse  "S3 presigner 失敗"
//	@Router       /ai-chat/attachments/upload-url [post]
//	@Security     CookieAuth
func (h *AiChatAttachmentHandler) IssueUploadURL(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.issueUseCase == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "attachment_upload_unavailable"})
		return
	}
	var body aiChatAttachmentUploadURLRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.ContentType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "contentType is required"})
		return
	}
	if body.Filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}
	if body.SizeBytes <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sizeBytes must be positive"})
		return
	}
	out, err := h.issueUseCase.Execute(c.Request.Context(), usecase.IssueAiChatAttachmentUploadURLInput{
		UserID:      uid,
		Filename:    body.Filename,
		ContentType: body.ContentType,
		SizeBytes:   body.SizeBytes,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrAttachmentUnsupportedType):
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrAttachmentTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		default:
			// 内部エラーは詳細を漏らさず 500。原因は server log にだけ残す。
			log.Printf("ai-chat-attachment: presigned URL issue failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.JSON(http.StatusOK, out)
}
