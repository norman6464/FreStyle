package handler

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase/profile"
)

// ProfileImageHandler は profile アイコン用 PUT 署名付き URL を発行する。
type ProfileImageHandler struct {
	issue *profile.IssueProfileImageUploadURLUseCase
}

func NewProfileImageHandler(i *profile.IssueProfileImageUploadURLUseCase) *ProfileImageHandler {
	return &ProfileImageHandler{issue: i}
}

// issueProfileImageReq は body 受け取り。fileName は受け取らない（オブジェクトの拡張子は
// 検査済みの contentType からのみ導く。rich_text_image_handler.go と同じ形）。
type issueProfileImageReq struct {
	ContentType string `json:"contentType"`
	// Size はバイト数。省略時は 0 になり、domain.ValidateImageUpload が「0 以下は拒否」で弾く。
	Size int64 `json:"size"`
}

var (
	errProfileImageForbidden    = errors.New("forbidden")
	errProfileImageUnauthorized = errors.New("unauthorized")
)

// resolveUserID は profile_handler と同じ規則で path :userId を解決する（"me" / 数字一致のみ通す）。
func (h *ProfileImageHandler) resolveUserID(c *gin.Context) (uint64, error) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		return 0, errProfileImageUnauthorized
	}
	param := c.Param("userId")
	if param == "" || param == "me" {
		return cur, nil
	}
	uid, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		//nolint:nilerr // 数字以外の userId は current user にフォールバックする設計（err は握り潰さず意図的に無視）
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errProfileImageForbidden
	}
	return uid, nil
}

// IssueUploadURL は { contentType, size } を受けて PUT 署名 URL 等を返す。
func (h *ProfileImageHandler) IssueUploadURL(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		switch {
		case errors.Is(err, errProfileImageUnauthorized):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errors.Is(err, errProfileImageForbidden):
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	var req issueProfileImageReq
	// body 無し（EOF）は許容するが、不正 JSON は 400 で弾く
	// （以前はここが全エラーを握り潰し、壊れた JSON でも既定値のまま処理を続けていた）。
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
