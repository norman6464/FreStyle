package handler

import (
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

// Create は POST /api/v2/admin/invitations。
//
// 認可ルール（SoD）:
//   - SuperAdmin: company_admin の招待のみ可能。trainee の招待は CompanyAdmin に任せる。
//     （運営が顧客企業の trainee 全員を直接管理するのは実運用に合わないため）
//   - CompanyAdmin: 自社の trainee の招待のみ可能。company_id は自社に固定する。
//   - その他: 403
//
// この境界を backend で確実に守ることで、フロント UI を経由しない API 呼び出しでも
// SuperAdmin が間違って trainee を直接招待することや、CompanyAdmin が他社に
// 招待を出すことを防ぐ。フロント UI は UX 改善として同じルールで選択肢を絞る。
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

	switch user.Role {
	case domain.RoleSuperAdmin:
		if req.Role != domain.RoleCompanyAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "super_admin_can_only_invite_company_admin",
				"message": "運営は会社管理者のみ招待できます。受講者の招待は会社管理者から行ってください。",
			})
			return
		}
	case domain.RoleCompanyAdmin:
		if req.Role != domain.RoleTrainee {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "company_admin_can_only_invite_trainee",
				"message": "会社管理者が招待できるのは受講者のみです。",
			})
			return
		}
		if user.CompanyID == nil || *user.CompanyID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "company_admin_without_company"})
			return
		}
		// CompanyAdmin の招待先 company は常に自社に固定する（リクエスト値を上書き）。
		req.CompanyID = *user.CompanyID
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	got, err := h.create.Execute(c.Request.Context(), usecase.CreateAdminInvitationInput{
		CompanyID: req.CompanyID, Email: req.Email, Role: req.Role, DisplayName: req.DisplayName,
	})
	if err != nil {
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
