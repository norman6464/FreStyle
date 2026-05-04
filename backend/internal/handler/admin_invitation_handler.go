package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	cognitoinfra "github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type AdminInvitationHandler struct {
	list   *usecase.ListAdminInvitationsUseCase
	create *usecase.CreateAdminInvitationUseCase
	cancel *usecase.CancelAdminInvitationUseCase
}

func NewAdminInvitationHandler(l *usecase.ListAdminInvitationsUseCase, c *usecase.CreateAdminInvitationUseCase, x *usecase.CancelAdminInvitationUseCase) *AdminInvitationHandler {
	return &AdminInvitationHandler{list: l, create: c, cancel: x}
}

// List は GET /api/v2/admin/invitations。
// 認可と scope 判定:
//   - SuperAdmin (運営): 全社横断 → ListAll() / ?companyId= 指定で絞り込みも可
//   - CompanyAdmin     : 自社の company_id で自動フィルタ → ListByCompanyID()
//   - その他           : 403 Forbidden
//
// 旧実装は ?companyId= クエリ必須だったが、フロントが渡していなかったため
// 常に 400 を返していた。current user の role / company_id から自動解決する設計に修正。
func (h *AdminInvitationHandler) List(c *gin.Context) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	switch user.Role {
	case domain.RoleSuperAdmin:
		// SuperAdmin は全社横断アクセス可。?companyId= が指定されていればそれで絞り込み。
		if q := c.Query("companyId"); q != "" {
			cid, _ := strconv.ParseUint(q, 10, 64)
			rows, err := h.list.ListByCompanyID(c.Request.Context(), cid)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, rows)
			return
		}
		rows, err := h.list.ListAll(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
			return
		}
		c.JSON(http.StatusOK, rows)
	case domain.RoleCompanyAdmin:
		// CompanyAdmin は自社のみ。company_id 未設定なら 403 (誤用防止)。
		if user.CompanyID == nil || *user.CompanyID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "company_admin_without_company"})
			return
		}
		rows, err := h.list.ListByCompanyID(c.Request.Context(), *user.CompanyID)
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
	CompanyID   uint64 `json:"companyId" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Role        string `json:"role" binding:"required"`
	DisplayName string `json:"displayName"`
}

func (h *AdminInvitationHandler) Create(c *gin.Context) {
	var req createAdminInvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateAdminInvitationInput{
		CompanyID: req.CompanyID, Email: req.Email, Role: req.Role, DisplayName: req.DisplayName,
	})
	if err != nil {
		if errors.Is(err, cognitoinfra.ErrUserAlreadyConfirmed) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "このメールアドレスはすでに登録済みです。再招待は不要です。",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *AdminInvitationHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.cancel.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
