package handler

import (
	"errors"
	"log"
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
// SizeBytes の上限は 5MB（5242880 byte / usecase.maxNoteImageBytes）。binding は正数であることのみを
// 保証し、上限超過の判定は usecase が行う（超過は 413）。
type issueUploadURLReq struct {
	ContentType string `json:"contentType" binding:"required"`
	SizeBytes   int64  `json:"sizeBytes" binding:"required,gt=0"`
}

// @Summary      ノート 画像 PUT 署名 URL
// @Description  current user 用 の S3 PUT 署名 URL を 発行。 contentType は 画像 MIME (png/jpeg/jpg/gif/webp) のみ、 sizeBytes は 上限 5MB を 事前 検証。 検証 済み の contentType / sizeBytes は presign の 署名 対象 (Content-Type / Content-Length) に 焼き込む ため、 発行後 に 別 種別 ・ 別 サイズ で PUT する と S3 が 署名 不一致 で 拒否 する。
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        body  body      issueUploadURLReq  true  "contentType / sizeBytes"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.NoteImageUploadURL
// @Failure      400   {object}  errorResponse  "Bad Request — contentType / sizeBytes が 未 指定、 sizeBytes が 正数 で ない、 または 不正 な JSON"
// @Failure      401   {object}  errorResponse  "Unauthorized — 未 認証 (Cookie の JWT が 無効 か 未 送信)"
// @Failure      413   {object}  errorResponse  "Payload Too Large — sizeBytes が 上限 5MB (5242880 byte) を 超えて いる。 5MB 以下 の 画像 を 選び直して 再送 する"
// @Failure      415   {object}  errorResponse  "Unsupported Media Type — contentType が 許可 された 画像 MIME (png/jpeg/jpg/gif/webp) で ない。 画像 以外 の ファイル は アップロード できない"
// @Failure      500   {object}  errorResponse  "presigned URL の 発行 に 失敗 (S3 / インフラ 側 の 異常)"
// @Router       /notes/images/upload-url [post]
// @Security     CookieAuth
func (h *NoteImageHandler) IssueUploadURL(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req issueUploadURLReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		case errors.Is(err, usecase.ErrNoteImageInvalidSize):
			// binding の gt=0 で通常は到達しないが、usecase 単体の契約として 400 に対応付ける。
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrNoteImageTooLarge):
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		default:
			// presigner のタイムアウト等はサーバ側の異常。400 で返すと監視がクライアント
			// エラーとして誤分類するため 500 とし、詳細は漏らさず server log にだけ残す。
			log.Printf("note-image: presigned URL issue failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.JSON(http.StatusOK, got)
}
