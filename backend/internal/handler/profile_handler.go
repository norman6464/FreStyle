package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// ProfileHandler は GET / PUT /profile/:userId(or "me") を提供する。
//
// 返却 DTO `domain.ProfileView` は users.display_name と profiles テーブルを合成したもの
// （フロントは displayName / bio / avatarUrl / status を 1 つの object で扱う）。
type ProfileHandler struct {
	get    *usecase.GetProfileUseCase
	update *usecase.UpdateProfileUseCase
	users  repository.UserRepository
}

func NewProfileHandler(
	g *usecase.GetProfileUseCase,
	u *usecase.UpdateProfileUseCase,
	users repository.UserRepository,
) *ProfileHandler {
	return &ProfileHandler{get: g, update: u, users: users}
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
	view, err := h.buildView(c, uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, view)
}

type updateProfileReq struct {
	// `name` は旧フロント実装の互換のため受け付け、`displayName` を優先する。
	DisplayName string `json:"displayName"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatarUrl"`
	// 旧フロント実装が `iconUrl` で送ってきた場合の互換。
	IconURL string `json:"iconUrl"`
	Status  string `json:"status"`
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
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}
	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = req.IconURL
	}
	if displayName != "" {
		if err := h.users.UpdateDisplayName(c.Request.Context(), uid, displayName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if _, err := h.update.Execute(c.Request.Context(), usecase.UpdateProfileInput{
		UserID:    uid,
		Bio:       req.Bio,
		AvatarURL: avatarURL,
		Status:    req.Status,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := h.buildView(c, uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": "プロフィールを更新しました"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// buildView は users.display_name と profiles を合成して ProfileView を返す。
// 未登録ユーザーでも UI が落ちないよう、欠損時は空文字フィールドで埋める。
func (h *ProfileHandler) buildView(c *gin.Context, uid uint64) (*domain.ProfileView, error) {
	p, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		return nil, err
	}
	view := &domain.ProfileView{UserID: uid}
	if p != nil {
		view.Bio = p.Bio
		view.AvatarURL = p.AvatarURL
		view.Status = p.Status
		view.UpdatedAt = p.UpdatedAt
	}
	user, _ := h.users.FindByID(c.Request.Context(), uid)
	if user != nil {
		view.DisplayName = user.DisplayName
	}
	return view, nil
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
