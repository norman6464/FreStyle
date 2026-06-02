package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// ProfileImageHandler は profile アイコン用 S3 PUT 署名付き URL を発行する。
type ProfileImageHandler struct {
	issue *usecase.IssueProfileImageUploadURLUseCase
}

func NewProfileImageHandler(i *usecase.IssueProfileImageUploadURLUseCase) *ProfileImageHandler {
	return &ProfileImageHandler{issue: i}
}

type issueProfileImageReq struct {
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
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
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errProfileImageForbidden
	}
	return uid, nil
}

// IssueUploadURL は { fileName, contentType } を受けて PUT 署名 URL 等を返す。
//
//	@Summary      プロフィール 画像 PUT 署名 URL
//	@Description  プロフィール アイコン 用 の S3 PUT 署名 URL を 発行。 "me" / 数字 一致 のみ 許可 (IDOR 対策)。 body は 任意。
//	@Tags         profile
//	@Accept       json
//	@Produce      json
//	@Param        userId  path      string                 true   "数字 ID または 'me'"
//	@Param        body    body      issueProfileImageReq   false  "fileName / contentType (任意)"
//	@Success      200     {object}  github_com_norman6464_FreStyle_backend_internal_domain.ProfileImageUploadURL
//	@Failure      400     {object}  errorResponse  "発行 失敗"
//	@Failure      401     {object}  errorResponse  "未 認証"
//	@Failure      403     {object}  errorResponse  "他 user 指定"
//	@Router       /profile/{userId}/image/presigned-url [post]
//	@Security     CookieAuth
func (h *ProfileImageHandler) IssueUploadURL(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		switch err {
		case errProfileImageUnauthorized:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errProfileImageForbidden:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	var req issueProfileImageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// body 無しで叩かれても 400 にせず、デフォルト値で処理を続ける。
		req = issueProfileImageReq{}
	}
	got, err := h.issue.Execute(c.Request.Context(), uid, req.FileName, req.ContentType)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}
