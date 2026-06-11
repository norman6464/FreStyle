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
// アカウントの有効/無効・論理削除を管理する。
type AdminMemberHandler struct {
	list       *usecase.ListCompanyMembersUseCase
	update     *usecase.UpdateMemberAiAccessUseCase
	setActive  *usecase.SetMemberActiveUseCase
	softDelete *usecase.SoftDeleteMemberUseCase
}

// NewAdminMemberHandler は一覧 / AI 更新 / 有効無効 / 論理削除 usecase を注入して handler を返す。
func NewAdminMemberHandler(
	list *usecase.ListCompanyMembersUseCase,
	update *usecase.UpdateMemberAiAccessUseCase,
	setActive *usecase.SetMemberActiveUseCase,
	softDelete *usecase.SoftDeleteMemberUseCase,
) *AdminMemberHandler {
	return &AdminMemberHandler{list: list, update: update, setActive: setActive, softDelete: softDelete}
}

// memberResponse は従業員一覧の 1 行（cognito_sub 等の機密は出さない）。
type memberResponse struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	// AiChatEnabled は AI 利用可否の個別上書き。null = 会社設定に従う。
	AiChatEnabled *bool `json:"aiChatEnabled"`
	// IsActive はアカウントの有効/無効。false = 無効（ログイン/利用不可）。
	IsActive bool `json:"isActive"`
}

func toMemberResponse(u domain.User) memberResponse {
	return memberResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		Role:          u.Role,
		AiChatEnabled: u.AiChatEnabled,
		IsActive:      u.IsActive,
	}
}

func isAdminActor(actor *domain.User) bool {
	return actor != nil && (actor.Role == domain.RoleCompanyAdmin || actor.Role == domain.RoleSuperAdmin)
}

// List は自社の従業員一覧を返す。
//
//	@Summary      従業員一覧
//	@Description  自社（company_admin の所属会社）の従業員一覧を返す。company_admin / super_admin のみ。
//	@Tags         admin
//	@Produce      json
//	@Success      200  {array}   memberResponse
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "管理者以外"
//	@Failure      500  {object}  errorResponse  "内部エラー"
//	@Router       /admin/members [get]
//	@Security     CookieAuth
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

type updateMemberAiRequest struct {
	// Enabled は AI 利用可否の個別上書き。null（未指定）で会社設定に従う状態へ戻す。
	Enabled *bool `json:"enabled"`
}

// UpdateAiAccess は自社の従業員の AI 利用可否を個別に更新する。
//
//	@Summary      従業員の AI 利用可否を個別更新
//	@Description  自社の従業員の AI 利用可否を個別に上書きする（null で会社設定に従う）。別会社の従業員は更新できない。company_admin / super_admin のみ。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        userId  path  string  true  "従業員の数値 ID"
//	@Param        body    body  updateMemberAiRequest  true  "enabled (null=会社設定に従う)"
//	@Success      204
//	@Failure      400  {object}  errorResponse  "バリデーション失敗"
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "管理者以外 / 別会社の従業員"
//	@Failure      500  {object}  errorResponse  "内部エラー"
//	@Router       /admin/members/{userId}/ai-access [patch]
//	@Security     CookieAuth
func (h *AdminMemberHandler) UpdateAiAccess(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isAdminActor(actor) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	var req updateMemberAiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.update.Execute(c.Request.Context(), actor, userID, req.Enabled); err != nil {
		if errors.Is(err, usecase.ErrMemberNotInActorCompany) {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
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
//
//	@Summary      従業員アカウントの有効/無効を切り替え
//	@Description  無効化すると、その従業員はログイン/利用不可になる（middleware で弾く）。super_admin は全社、company_admin は自社の従業員のみ。自分自身は不可。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        userId  path  string  true  "従業員の数値 ID"
//	@Param        body    body  setMemberActiveRequest  true  "active=false で無効化"
//	@Success      204
//	@Failure      400  {object}  errorResponse  "不正な ID / body / 自分自身"
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "管理者以外 / 別会社の従業員"
//	@Failure      404  {object}  errorResponse  "従業員が存在しない"
//	@Failure      500  {object}  errorResponse  "内部エラー"
//	@Router       /admin/members/{userId}/active [patch]
//	@Security     CookieAuth
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
//
//	@Summary      従業員を論理削除
//	@Description  従業員を一覧から退会させる（論理削除）。以後ログイン/利用不可。super_admin は全社、company_admin は自社の従業員のみ。自分自身は不可。
//	@Tags         admin
//	@Produce      json
//	@Param        userId  path  string  true  "従業員の数値 ID"
//	@Success      204
//	@Failure      400  {object}  errorResponse  "不正な ID / 自分自身"
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "管理者以外 / 別会社の従業員"
//	@Failure      404  {object}  errorResponse  "従業員が存在しない"
//	@Failure      500  {object}  errorResponse  "内部エラー"
//	@Router       /admin/members/{userId} [delete]
//	@Security     CookieAuth
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
