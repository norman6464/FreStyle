package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AdminMemberHandler は company_admin / super_admin が従業員一覧と、各従業員の AI 利用可否・
// アカウントの有効/無効・論理削除・学習状況サマリーを扱う。
type AdminMemberHandler struct {
	list            *usecase.ListCompanyMembersUseCase
	setActive       *usecase.SetMemberActiveUseCase
	softDelete      *usecase.SoftDeleteMemberUseCase
	learningSummary *usecase.GetCompanyLearningSummaryUseCase
}

// NewAdminMemberHandler は一覧 / 有効無効 / 論理削除 / 学習サマリー usecase を注入して handler を返す。
func NewAdminMemberHandler(
	list *usecase.ListCompanyMembersUseCase,
	setActive *usecase.SetMemberActiveUseCase,
	softDelete *usecase.SoftDeleteMemberUseCase,
	learningSummary *usecase.GetCompanyLearningSummaryUseCase,
) *AdminMemberHandler {
	return &AdminMemberHandler{
		list:            list,
		setActive:       setActive,
		softDelete:      softDelete,
		learningSummary: learningSummary,
	}
}

// memberResponse は従業員一覧の 1 行（cognito_sub 等の機密は出さない）。
type memberResponse struct {
	ID    uint64          `json:"id"`
	Email string          `json:"email"`
	Name  string          `json:"name"`
	Role  domain.RoleName `json:"role"`
	// IsActive はアカウントの有効/無効。false = 無効（ログイン/利用不可）。
	IsActive bool `json:"isActive"`
}

func toMemberResponse(u domain.User) memberResponse {
	return memberResponse{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Role:     u.Role,
		IsActive: u.IsActive,
	}
}

func isAdminActor(actor *domain.User) bool {
	return actor != nil && (actor.Role == domain.RoleCompanyAdmin || actor.Role == domain.RoleSuperAdmin)
}

// List は自社の従業員一覧を返す。
func (h *AdminMemberHandler) List(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isAdminActor(actor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	members, err := h.list.Execute(c.Request.Context(), actor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	out := make([]memberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, toMemberResponse(m))
	}
	c.JSON(http.StatusOK, out)
}

// LearningSummary は自社 trainee の学習状況サマリーを返す(company_admin のホーム用)。
func (h *AdminMemberHandler) LearningSummary(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isAdminActor(actor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	summary, err := h.learningSummary.Execute(c.Request.Context(), actor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// memberOpErrorStatus は従業員の停止/削除 usecase のエラーを HTTP ステータスにマップする。
func memberOpErrorStatus(err error) (int, string) {
	switch {
	case errors.Is(err, usecase.ErrMemberNotFound):
		return http.StatusNotFound, "member_not_found"
	case errors.Is(err, usecase.ErrCannotManageSelf):
		return http.StatusBadRequest, "cannot_manage_self"
	case errors.Is(err, usecase.ErrMemberNotInActorCompany):
		return http.StatusForbidden, "forbidden"
	default:
		return http.StatusInternalServerError, "internal_error"
	}
}

// setMemberActiveRequest は従業員アカウントの有効/無効更新の入力。
type setMemberActiveRequest struct {
	Active *bool `json:"active" binding:"required"`
}

// SetActive は従業員アカウントの有効/無効を切り替える（停止/再開）。
func (h *AdminMemberHandler) SetActive(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isAdminActor(actor) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_user_id"})
		return
	}
	var req setMemberActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	if err := h.setActive.Execute(c.Request.Context(), actor, userID, *req.Active); err != nil {
		code, msg := memberOpErrorStatus(err)
		c.JSON(code, errorResponse{Error: msg})
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete は従業員を論理削除する（deleted_at = NOW()）。
func (h *AdminMemberHandler) Delete(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isAdminActor(actor) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_user_id"})
		return
	}
	if err := h.softDelete.Execute(c.Request.Context(), actor, userID); err != nil {
		code, msg := memberOpErrorStatus(err)
		c.JSON(code, errorResponse{Error: msg})
		return
	}
	c.Status(http.StatusNoContent)
}
