package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type AdminInvitationHandler struct {
	list       *usecase.ListAdminInvitationsUseCase
	create     *usecase.CreateAdminInvitationUseCase
	tempCreate *usecase.CreateTemporaryPasswordInvitationUseCase
	cancel     *usecase.CancelAdminInvitationUseCase
}

func NewAdminInvitationHandler(
	l *usecase.ListAdminInvitationsUseCase,
	c *usecase.CreateAdminInvitationUseCase,
	t *usecase.CreateTemporaryPasswordInvitationUseCase,
	x *usecase.CancelAdminInvitationUseCase,
) *AdminInvitationHandler {
	return &AdminInvitationHandler{list: l, create: c, tempCreate: t, cancel: x}
}

// List は招待一覧を返す。SuperAdmin は全社横断（?companyId= で絞り込み可）、
// CompanyAdmin は自社のみ、それ以外は 403。
//
//	@Summary      招待 一覧 (admin)
//	@Description  pending な 招待 を 返す。 SuperAdmin は 全社 (?companyId= で 絞り込み 可)、 CompanyAdmin は 自社 のみ。 trainee 等 は 403。
//	@Tags         admin
//	@Produce      json
//	@Param        companyId  query     string  false  "SuperAdmin の とき のみ 有効: 特定 company の 招待 のみ"
//	@Success      200        {array}   github_com_norman6464_FreStyle_backend_internal_domain.AdminInvitation
//	@Failure      400        {object}  errorResponse  "ListByCompanyID 失敗 (現状 実装 で 400 を 返す パス あり)"
//	@Failure      401        {object}  errorResponse  "未 認証"
//	@Failure      403        {object}  errorResponse  "trainee / company 未 設定 等"
//	@Failure      500        {object}  errorResponse  "DB 失敗 (ListAll 経路)"
//	@Router       /admin/invitations [get]
//	@Security     CookieAuth
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
	CompanyID uint64          `json:"companyId" binding:"required"`
	Email     string          `json:"email" binding:"required"`
	Role      domain.RoleName `json:"role" binding:"required"`
	Name      string          `json:"name"`
	// Method は招待方式。"magic_link"（既定・受諾リンクをメール）か
	// "temporary_password"（Cognito 一時パスワードを発行し 1 度だけ返す・FRESTYLE-313）。
	Method string `json:"method"`
}

// 招待方式の値。
const (
	invMethodMagicLink = "magic_link"
	invMethodTempPass  = "temporary_password"
)

// Create は招待を作成する。SoD: SuperAdmin は company_admin のみ、CompanyAdmin は自社の trainee のみ招待可。
// この境界は backend で確実に守り、UI を経由しない呼び出しでも越権招待を防ぐ。
//
//	@Summary      招待 作成
//	@Description  招待を作成する。method=magic_link（既定）は受諾リンクをメール送信、method=temporary_password は Cognito 一時パスワードを発行してレスポンスで 1 度だけ返す。SoD: SuperAdmin は company_admin のみ 招待 可、 CompanyAdmin は trainee のみ 自社 に 招待 可。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        body  body      createAdminInvReq  true  "招待 内容 (CompanyAdmin は companyId が 上書き さ れる)"
//	@Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.AdminInvitation  "magic_link 方式は招待行。temporary_password 方式は {invitation, temporaryPassword} を返し temporaryPassword は 1 度だけ提示される"
//	@Failure      400   {object}  errorResponse  "バリデーション / 未知の method / 一時パスワード方式が未構成"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      403   {object}  errorResponse  "ロール 違反"
//	@Failure      409   {object}  errorResponse  "一時パスワード方式で対象 email が既に存在"
//	@Router       /admin/invitations [post]
//	@Security     CookieAuth
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
	// 未知の method は黙ってマジックリンクにフォールバックさせない（意図しない招待メール送信を防ぐ）。
	switch req.Method {
	case "", invMethodMagicLink, invMethodTempPass:
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_method"})
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

	in := usecase.CreateAdminInvitationInput{
		CompanyID: req.CompanyID, Email: req.Email, Role: req.Role, Name: req.Name,
	}

	if req.Method == invMethodTempPass {
		h.createWithTemporaryPassword(c, in)
		return
	}

	// 既定はマジックリンク方式。
	got, err := h.create.Execute(c.Request.Context(), in)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

// createWithTemporaryPassword は初期パスワード方式の招待を作り、一時パスワードを 1 度だけ返す。
// 一時パスワードはレスポンスにのみ含め、保存・ログ出力しない（FRESTYLE-313）。
func (h *AdminInvitationHandler) createWithTemporaryPassword(c *gin.Context, in usecase.CreateAdminInvitationInput) {
	if h.tempCreate == nil {
		// usecase の ErrTemporaryPasswordUnavailable と同じく 400 に統一（未構成 = 提供していない方式）。
		c.JSON(http.StatusBadRequest, gin.H{"error": "temporary_password_not_configured"})
		return
	}
	out, err := h.tempCreate.Execute(c.Request.Context(), in)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrTemporaryPasswordUnavailable):
			c.JSON(http.StatusBadRequest, gin.H{"error": "temporary_password_not_configured"})
		case errors.Is(err, cognito.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "user_already_exists",
				"message": "この email のユーザーは既に存在します。パスワードの再発行は別途行ってください。",
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"invitation":        out.Invitation,
		"temporaryPassword": out.TemporaryPassword,
	})
}

// @Summary      招待 取り消し
// @Description  指定 招待 の status を canceled に 更新。 行 は 物理 削除 せず 監査 用 に 残す。
// @Description  super_admin は 全社、 company_admin は 自社 の 招待 のみ 取消 できる。
// @Tags         admin
// @Produce      json
// @Param        id  path  int  true  "招待 ID"
// @Success      204  "成功 (本文 なし)"
// @Failure      400  {object}  errorResponse  "不正 な ID / DB 失敗"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "管理者 以外"
// @Failure      404  {object}  errorResponse  "招待 が 存在 し ない (他社 の 招待 を 含む)"
// @Router       /admin/invitations/{id} [delete]
// @Security     CookieAuth
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

	// CompanyID は SuperAdmin では nil になり得る（usecase 側で role により無視される）。
	var actorCompanyID uint64
	if user.CompanyID != nil {
		actorCompanyID = *user.CompanyID
	}

	err = h.cancel.Execute(c.Request.Context(), usecase.CancelAdminInvitationInput{
		ID:             id,
		ActorRole:      user.Role,
		ActorCompanyID: actorCompanyID,
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
