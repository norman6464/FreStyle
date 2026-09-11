package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/profile"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ProfileHandler は GET / PUT /profile/:userId(or "me") を提供する。
// 返却する domain.ProfileView は users.name と profiles を合成したもの。
type ProfileHandler struct {
	get    *profile.GetProfileUseCase
	update *profile.UpdateProfileUseCase
	users  repository.UserRepository
}

func NewProfileHandler(
	g *profile.GetProfileUseCase,
	u *profile.UpdateProfileUseCase,
	users repository.UserRepository,
) *ProfileHandler {
	return &ProfileHandler{get: g, update: u, users: users}
}

var (
	errProfileForbidden    = errors.New("forbidden")
	errProfileUnauthorized = errors.New("unauthorized")
)

// resolveUserID は "me" / 空文字 / 数字以外を current user に、数字一致はそのまま、
// 数字で current user 以外は 403 にする（IDOR 対策）。
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
		//nolint:nilerr // 数字以外の userId は current user にフォールバックする設計（err は握り潰さず意図的に無視）
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errProfileForbidden
	}
	return uid, nil
}

// Get は指定 user のプロフィールを返す。
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

// 各項目の上限は、本文サイズの全体上限（middleware.MaxRequestBody）とは別に、
// 1 項目だけが極端に大きい値で DB へ届くのを入口で弾くためのもの
// （DB の列自体は text で無制限。ここで切らないと、表示側が想定しない長さの値に
// 対処し続けることになる）。
type updateProfileReq struct {
	Name      string `json:"displayName" binding:"omitempty,max=200"`
	Bio       string `json:"bio"         binding:"omitempty,max=2000"`
	AvatarURL string `json:"avatarUrl"   binding:"omitempty,max=2000"`
	IconURL   string `json:"iconUrl"     binding:"omitempty,max=2000"` // 旧フロント互換。avatarUrl を優先。
	Status    string `json:"status"      binding:"omitempty,max=200"`
}

// Update は current user のプロフィールを更新する。
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
	name := req.Name
	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = req.IconURL
	}
	if name != "" {
		if err := h.users.UpdateName(c.Request.Context(), uid, name); err != nil {
			// 1 行も更新できなかった（= リクエスト中に user 行が消えた）。以前は 0 件でも
			// 成功扱いで 200 を返しており、氏名が保存されていないのに保存済みに見えていた。
			if errors.Is(err, domain.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if _, err := h.update.Execute(c.Request.Context(), profile.UpdateProfileInput{
		UserID:        uid,
		Bio:           req.Bio,
		AvatarURL:     avatarURL,
		StatusMessage: req.Status,
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

// buildView は users.name と profiles を合成して ProfileView を返す（欠損時は空文字で埋める）。
func (h *ProfileHandler) buildView(c *gin.Context, uid uint64) (*domain.ProfileView, error) {
	p, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		return nil, err
	}
	view := &domain.ProfileView{UserID: uid}
	if p != nil {
		view.Bio = p.Bio
		view.AvatarURL = p.AvatarURL
		view.StatusMessage = p.StatusMessage
		view.UpdatedAt = p.UpdatedAt
	}
	user, _ := h.users.FindByID(c.Request.Context(), uid)
	if user != nil {
		view.Name = user.Name
		view.Email = user.Email
	}
	return view, nil
}

func writeProfileError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, errProfileUnauthorized):
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	case errors.Is(err, errProfileForbidden):
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
