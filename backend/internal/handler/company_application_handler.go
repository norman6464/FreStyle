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

// CompanyApplicationHandler は企業利用申請の作成（公開）と一覧 / status 更新（super_admin）を扱う。
type CompanyApplicationHandler struct {
	create       *usecase.CreateCompanyApplicationUseCase
	list         *usecase.ListCompanyApplicationsUseCase
	updateStatus *usecase.UpdateCompanyApplicationStatusUseCase
}

func NewCompanyApplicationHandler(
	create *usecase.CreateCompanyApplicationUseCase,
	list *usecase.ListCompanyApplicationsUseCase,
	updateStatus *usecase.UpdateCompanyApplicationStatusUseCase,
) *CompanyApplicationHandler {
	return &CompanyApplicationHandler{create: create, list: list, updateStatus: updateStatus}
}

type createCompanyApplicationReq struct {
	CompanyName   string `json:"companyName" binding:"required"`
	ApplicantName string `json:"applicantName" binding:"required"`
	Email         string `json:"email" binding:"required"`
	Message       string `json:"message"`
}

// Create は未登録ユーザーが出す企業申請を受け付ける（認証不要）。
//
//	@Summary      企業利用申請（公開 / 認証不要）
//	@Description  ログイン前のユーザーが会社名 / 氏名 / メール / 任意メッセージで利用申請を送る。受理時に super_admin へ通知する。
//	@Tags         company-applications
//	@Accept       json
//	@Produce      json
//	@Param        body  body      createCompanyApplicationReq  true  "申請内容"
//	@Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.CompanyApplication
//	@Failure      400   {object}  errorResponse  "バリデーションエラー"
//	@Failure      500   {object}  errorResponse  "内部エラー"
//	@Router       /company-applications [post]
func (h *CompanyApplicationHandler) Create(c *gin.Context) {
	var req createCompanyApplicationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	app, err := h.create.Execute(c.Request.Context(), usecase.CreateCompanyApplicationInput{
		CompanyName:   req.CompanyName,
		ApplicantName: req.ApplicantName,
		Email:         req.Email,
		Message:       req.Message,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrCompanyApplicationInvalid) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusCreated, app)
}

// List は全企業申請を返す（super_admin 専用）。
//
//	@Summary      企業申請一覧（super_admin）
//	@Description  受け付けた企業申請を新しい順で返す。super_admin 専用。
//	@Tags         admin
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.CompanyApplication
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "super_admin 以外"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /admin/company-applications [get]
//	@Security     CookieAuth
func (h *CompanyApplicationHandler) List(c *gin.Context) {
	if !h.requireSuperAdmin(c) {
		return
	}
	rows, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type updateCompanyApplicationStatusReq struct {
	Status string `json:"status" binding:"required"`
}

// UpdateStatus は申請の status を更新する（super_admin 専用）。
//
//	@Summary      企業申請の status 更新（super_admin）
//	@Description  申請を approved / rejected / pending に更新する。super_admin 専用。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        id    path      int                                 true  "申請 ID"
//	@Param        body  body      updateCompanyApplicationStatusReq   true  "status"
//	@Success      204   "成功（本文なし）"
//	@Failure      400   {object}  errorResponse  "id / status 不正"
//	@Failure      401   {object}  errorResponse  "未認証"
//	@Failure      403   {object}  errorResponse  "super_admin 以外"
//	@Failure      500   {object}  errorResponse  "DB 更新失敗"
//	@Router       /admin/company-applications/{id}/status [patch]
//	@Security     CookieAuth
func (h *CompanyApplicationHandler) UpdateStatus(c *gin.Context) {
	if !h.requireSuperAdmin(c) {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateCompanyApplicationStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.updateStatus.Execute(c.Request.Context(), id, req.Status); err != nil {
		// status 不正のみ 400。DB 更新失敗等の内部エラーは詳細を漏らさず 500。
		if errors.Is(err, usecase.ErrCompanyApplicationBadStatus) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.Status(http.StatusNoContent)
}

// requireSuperAdmin は current user が super_admin でなければ 401/403 を返して false を返す。
func (h *CompanyApplicationHandler) requireSuperAdmin(c *gin.Context) bool {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return false
	}
	if user.Role != domain.RoleSuperAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return false
	}
	return true
}
