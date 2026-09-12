package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/profile"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ProfileHandler は GET / PUT /profile/:userId(or "me")、PUT /me/status、
// GET /me/identities を提供する。返却する domain.ProfileView は users.name と
// profiles を合成したもの。
type ProfileHandler struct {
	get            *profile.GetProfileUseCase
	update         *profile.UpdateProfileUseCase
	updateStatus   *profile.UpdateStatusUseCase
	listIdentities *profile.ListMyIdentitiesUseCase
	users          repository.UserRepository
}

func NewProfileHandler(
	g *profile.GetProfileUseCase,
	u *profile.UpdateProfileUseCase,
	updateStatus *profile.UpdateStatusUseCase,
	listIdentities *profile.ListMyIdentitiesUseCase,
	users repository.UserRepository,
) *ProfileHandler {
	return &ProfileHandler{get: g, update: u, updateStatus: updateStatus, listIdentities: listIdentities, users: users}
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
		//nolint:nilerr // 数字以外の userId は current user にフォールバックする設計（意図的に無視）
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

// 各項目の上限は本文サイズの全体上限とは別に、1 項目だけ極端に大きい値が DB へ届くのを
// 入口で弾くためのもの（DB の列自体は text で無制限）。
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
			// 1 行も更新できなかった（リクエスト中に user 行が消えた）。保存されていないのに
			// 保存済みに見せないよう、0 件更新は成功扱いにしない。
			if errors.Is(err, domain.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if _, err := h.update.Execute(c.Request.Context(), profile.UpdateProfileInput{
		UserID:     uid,
		Bio:        req.Bio,
		AvatarURL:  avatarURL,
		StatusText: req.Status,
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

// Emoji の上限は結合絵文字（ZWJ シーケンス等）が単一の絵文字でも複数バイトになり得るため、
// 普通の一言テキストより広めに取る。
type updateStatusReq struct {
	Emoji     string     `json:"emoji"     binding:"omitempty,max=32"`
	Text      string     `json:"text"      binding:"omitempty,max=200"`
	ExpiresAt *time.Time `json:"expiresAt"` // nil/省略 = 無期限
}

// UpdateStatus は一言ステータス（絵文字・テキスト・失効時刻）だけを更新する（PUT /me/status）。
// bio / avatarUrl には触れない（Update の専管）。
func (h *ProfileHandler) UpdateStatus(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	var req updateStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := h.updateStatus.Execute(c.Request.Context(), profile.UpdateStatusInput{
		UserID:    uid,
		Emoji:     req.Emoji,
		Text:      req.Text,
		ExpiresAt: req.ExpiresAt,
	}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := h.buildView(c, uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": "ステータスを更新しました"})
		return
	}
	c.JSON(http.StatusOK, view)
}

// profileIdentityResponse は認証方法 1 件の返却形（表示専用）。
type profileIdentityResponse struct {
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"createdAt"`
}

// ListIdentities は本人の認証方法一覧を返す（GET /me/identities）。resolveUserID が現在
// ユーザー以外の数値 userId を 403 にするので、他人の subject はこの経路では読めない。
func (h *ProfileHandler) ListIdentities(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		writeProfileError(c, err)
		return
	}
	identities, err := h.listIdentities.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out := make([]profileIdentityResponse, 0, len(identities))
	for _, id := range identities {
		out = append(out, profileIdentityResponse{Provider: id.Provider, Subject: id.Subject, CreatedAt: id.CreatedAt})
	}
	c.JSON(http.StatusOK, out)
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
		view.StatusText = p.StatusText
		view.StatusEmoji = p.StatusEmoji
		view.StatusExpiresAt = p.StatusExpiresAt
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
