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

// AdminMemberHandler は company_admin が自社の従業員一覧と、各従業員の AI 利用可否を管理する。
type AdminMemberHandler struct {
	list   *usecase.ListCompanyMembersUseCase
	update *usecase.UpdateMemberAiAccessUseCase
}

// NewAdminMemberHandler は一覧 / AI 更新 usecase を注入して handler を返す。
func NewAdminMemberHandler(list *usecase.ListCompanyMembersUseCase, update *usecase.UpdateMemberAiAccessUseCase) *AdminMemberHandler {
	return &AdminMemberHandler{list: list, update: update}
}

// memberResponse は従業員一覧の 1 行（cognito_sub 等の機密は出さない）。
type memberResponse struct {
	ID          uint64 `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	// AiChatEnabled は AI 利用可否の個別上書き。null = 会社設定に従う。
	AiChatEnabled *bool `json:"aiChatEnabled"`
}

func toMemberResponse(u domain.User) memberResponse {
	return memberResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		Role:          u.Role,
		AiChatEnabled: u.AiChatEnabled,
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
