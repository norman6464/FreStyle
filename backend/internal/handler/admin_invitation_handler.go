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

type AdminInvitationHandler struct {
	list   *usecase.ListAdminInvitationsUseCase
	create *usecase.CreateAdminInvitationUseCase
	cancel *usecase.CancelAdminInvitationUseCase
}

func NewAdminInvitationHandler(
	l *usecase.ListAdminInvitationsUseCase,
	c *usecase.CreateAdminInvitationUseCase,
	x *usecase.CancelAdminInvitationUseCase,
) *AdminInvitationHandler {
	return &AdminInvitationHandler{list: l, create: c, cancel: x}
}

// List は招待一覧を返す。SuperAdmin は横断、CompanyAdmin は自分の所属のみ、それ以外は 403。
func (h *AdminInvitationHandler) List(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	switch user.Role {
	case domain.RoleSuperAdmin:
		rows, err := h.list.ListAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		c.JSON(http.StatusOK, rows)
	case domain.RoleCompanyAdmin:
		// CompanyAdmin は自分の所属のみ。所属ワークスペースが無ければ絞り込み先が
		// 決まらないので 403 (誤用防止)。
		workspaceID, affiliated := user.WorkspaceRef().WorkspaceID()
		if !affiliated {
			c.JSON(http.StatusForbidden, gin.H{"error": "company_admin_without_company"})
			return
		}
		rows, err := h.list.ListByWorkspaceID(c.Request.Context(), workspaceID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, rows)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	}
}

type createAdminInvReq struct {
	Email string          `json:"email" binding:"required"`
	Role  domain.RoleName `json:"role" binding:"required"`
	Name  string          `json:"name"`
	// Method は招待方式。いまは "magic_link"（受諾リンクをメールで送る）だけ。
	// 未知の値は binding で 400 にする（黙って別の方式にフォールバックさせない）。
	//
	// 初期パスワード方式は撤去した。発行者側にユーザーを作って一時パスワードを配る
	// 仕組みで、特定の発行者の管理 API に直結していた。ログインが発行者の画面に
	// 移った今、アプリが人のパスワードを決めて渡す経路そのものを持たない。
	Method string `json:"method" binding:"omitempty,oneof=magic_link"`
}

// Create は招待を作成する。招待先は actor 自身の所属ワークスペースに固定され、招待できるのは trainee のみ。
// この境界は backend で確実に守り、UI を経由しない呼び出しでも越権招待を防ぐ。
func (h *AdminInvitationHandler) Create(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createAdminInvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 招待先は常に actor 自身の所属ワークスペースに固定する。テナントを横断して
	// 招ける中央管理者は置かないので、招待先をリクエストで指定させる入口も無い。
	if user.Role != domain.RoleCompanyAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	if req.Role != domain.RoleTrainee {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "company_admin_can_only_invite_trainee",
			"message": "ワークスペース管理者が招待できるのは受講者のみです。",
		})
		return
	}
	// 招待先が決まらないので、未所属の管理者は招待できない。
	if _, affiliated := user.WorkspaceRef().WorkspaceID(); !affiliated {
		c.JSON(http.StatusForbidden, gin.H{"error": "company_admin_without_company"})
		return
	}

	in := usecase.CreateAdminInvitationInput{
		TargetWorkspace: user.WorkspaceRef(),
		Email:           req.Email,
		Role:            req.Role,
		Name:            req.Name,
	}

	got, err := h.create.Execute(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *AdminInvitationHandler) Cancel(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id"})
		return
	}

	// 所属ワークスペースは SuperAdmin では未所属になり得る（usecase 側で role により無視される）。
	err = h.cancel.Execute(c.Request.Context(), usecase.CancelAdminInvitationInput{
		ID:             id,
		ActorRole:      user.Role,
		ActorWorkspace: user.WorkspaceRef(),
	})
	switch {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, usecase.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, usecase.ErrInvitationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
