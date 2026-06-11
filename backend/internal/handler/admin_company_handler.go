package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// isSuperAdmin は actor が super_admin（運営管理者）かを判定する。
func isSuperAdmin(actor *domain.User) bool {
	return actor != nil && actor.Role == domain.RoleSuperAdmin
}

// AdminCompanyHandler は super_admin 向けの会社一覧と、会社アカウントの有効/無効を扱う。
type AdminCompanyHandler struct {
	list      *usecase.ListCompaniesUseCase
	setActive *usecase.SetCompanyActiveUseCase
}

func NewAdminCompanyHandler(l *usecase.ListCompaniesUseCase, s *usecase.SetCompanyActiveUseCase) *AdminCompanyHandler {
	return &AdminCompanyHandler{list: l, setActive: s}
}

// @Summary      会社 一覧 (SuperAdmin)
// @Description  全 company を 返す。 super_admin 専用 画面 用。 認可 は middleware で 別途 担保。
// @Tags         admin
// @Produce      json
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.Company
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /admin/companies [get]
// @Security     CookieAuth
func (h *AdminCompanyHandler) List(c *gin.Context) {
	companies, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, companies)
}

// setCompanyActiveRequest は会社アカウントの有効/無効更新の入力。
type setCompanyActiveRequest struct {
	Active *bool `json:"active" binding:"required"`
}

// SetActive は会社アカウントの有効/無効を切り替えるハンドラ（super_admin 専用）。
//
//	@Summary      会社アカウントの有効/無効を切り替え（super_admin 専用）
//	@Description  会社を無効化すると、その会社の全ユーザーがログイン/利用不可になる（middleware で弾く）。super_admin のみ。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        id    path      int                      true  "会社 ID"
//	@Param        body  body      setCompanyActiveRequest  true  "active=false で無効化"
//	@Success      200   {object}  messageResponse
//	@Failure      400   {object}  errorResponse  "不正な ID / body"
//	@Failure      401   {object}  errorResponse  "未認証"
//	@Failure      403   {object}  errorResponse  "super_admin 以外"
//	@Failure      404   {object}  errorResponse  "会社が存在しない"
//	@Failure      500   {object}  errorResponse  "更新失敗"
//	@Router       /admin/companies/{id}/active [patch]
//	@Security     CookieAuth
func (h *AdminCompanyHandler) SetActive(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isSuperAdmin(actor) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_id"})
		return
	}

	var body setCompanyActiveRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}

	err = h.setActive.Execute(c.Request.Context(), usecase.SetCompanyActiveInput{
		CompanyID: id,
		Active:    *body.Active,
	})
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "company_not_found"})
		return
	case err != nil:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "update_failed"})
		return
	}

	c.JSON(http.StatusOK, messageResponse{Message: "会社の状態を更新しました"})
}
