package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ProfileHandler struct {
	get    *usecase.GetProfileUseCase
	update *usecase.UpdateProfileUseCase
}

func NewProfileHandler(g *usecase.GetProfileUseCase, u *usecase.UpdateProfileUseCase) *ProfileHandler {
	return &ProfileHandler{get: g, update: u}
}

var (
	errProfileForbidden    = errors.New("forbidden")
	errProfileUnauthorized = errors.New("unauthorized")
)

// resolveUserID は path :userId を current user と突き合わせる。
//   - "me" / 空文字 / 数字以外 → current user に解決（フロント /profile/me に対応）
//   - 数字で current user と一致 → そのまま
//   - 数字で current user 以外 → 403（IDOR 対策）
func (h *ProfileHandler) resolveUserID(c *gin.Context) (uint64, error) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		return 0, errProfileUnauthorized
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
		return 0, errProfileForbidden
	}
	return uid, nil
}

func (h *ProfileHandler) Get(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	p, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		// 未登録ユーザーでも UI が落ちないよう最小デフォルトを返す。
		c.JSON(http.StatusOK, gin.H{"userId": uid, "bio": "", "avatarUrl": "", "name": "", "iconUrl": "", "status": ""})
		return
	}
	c.JSON(http.StatusOK, p)
}

type updateProfileReq struct {
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatarUrl"`
}

func (h *ProfileHandler) Update(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.update.Execute(c.Request.Context(), usecase.UpdateProfileInput{
		UserID: uid, Bio: req.Bio, AvatarURL: req.AvatarURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func writeProfileError(c *gin.Context, err error) {
	switch err {
	case errProfileUnauthorized:
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errProfileForbidden:
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
