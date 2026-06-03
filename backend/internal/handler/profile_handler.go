package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ProfileHandler は GET / PUT /profile/:userId(or "me") を提供する。
// 返却する domain.ProfileView は users.display_name と profiles を合成したもの。
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
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errProfileForbidden
	}
	return uid, nil
}

// Get は指定 user のプロフィールを返す。
//
//	@Summary      プロフィール 取得
//	@Description  指定 user (or current user) の displayName / bio / avatarUrl / status を 返す。 IDOR 対策 で 自分 以外 は 403。
//	@Tags         profile
//	@Produce      json
//	@Param        userId  path      string  true   "数字 ID または 'me'"
//	@Success      200     {object}  domain.ProfileView
//	@Failure      400     {object}  errorResponse  "view 構築 失敗 (profile / users 取得 の 内部 エラー 等)"
//	@Failure      401     {object}  errorResponse  "未 認証"
//	@Failure      403     {object}  errorResponse  "他 user の userId 指定"
//	@Router       /profile/{userId} [get]
//	@Security     CookieAuth
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
	DisplayName string `json:"displayName"`
	Name        string `json:"name"` // 旧フロント互換。displayName を優先。
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatarUrl"`
	IconURL     string `json:"iconUrl"` // 旧フロント互換。avatarUrl を優先。
	Status      string `json:"status"`
}

// Update は current user のプロフィールを更新する。
//
//	@Summary      プロフィール 更新
//	@Description  current user の displayName / bio / avatarUrl / status を 更新 する。 他 user は 403。
//	@Tags         profile
//	@Accept       json
//	@Produce      json
//	@Param        userId  path      string            true  "数字 ID または 'me'"
//	@Param        body    body      updateProfileReq  true  "更新 内容 (任意 フィールド の み)"
//	@Success      200     {object}  domain.ProfileView
//	@Failure      400     {object}  errorResponse  "バリデーション エラー"
//	@Failure      401     {object}  errorResponse  "未 認証"
//	@Failure      403     {object}  errorResponse  "他 user 指定"
//	@Router       /profile/{userId} [put]
//	@Security     CookieAuth
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

// buildView は users.display_name と profiles を合成して ProfileView を返す（欠損時は空文字で埋める）。
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
