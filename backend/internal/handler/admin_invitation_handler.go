package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
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
//	@Failure      400        {object}  errorResponse  "会社指定の一覧取得 失敗 (現状 実装 で 400 を 返す パス あり)"
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
		// CompanyAdmin は自社のみ。所属ワークスペースが無ければ絞り込み先が決まらないので
		// 403 (誤用防止)。SuperAdmin の ?companyId= 絞り込みは usecase が
		// company → workspace へ読み替えてから引く。
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
	CompanyID uint64          `json:"companyId" binding:"required"`
	Email     string          `json:"email" binding:"required"`
	Role      domain.RoleName `json:"role" binding:"required"`
	Name      string          `json:"name"`
	// Method は招待方式。"magic_link"（既定・受諾リンクをメール）か
	// "temporary_password"（Cognito 一時パスワードを発行し 1 度だけ返す・FRESTYLE-313）。
	// 未知の値は binding で 400 にする（黙ってマジックリンクにフォールバックさせない）。
	Method string `json:"method" binding:"omitempty,oneof=magic_link temporary_password"`
}

// tempPasswordInvitationResponse は temporary_password 方式のレスポンス。
// temporaryPassword は 1 度だけ提示され、保存・再取得はできない（FRESTYLE-313）。
type tempPasswordInvitationResponse struct {
	Invitation        *domain.AdminInvitation `json:"invitation"`
	TemporaryPassword string                  `json:"temporaryPassword"`
}

// 招待方式の値（temporary_password のみコードで分岐。magic_link は binding の既定/明示両対応）。
const invMethodTempPass = "temporary_password"

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

	// CompanyAdmin 経路だけ「招待先 = actor 自身の所属」が確定しているので、会社を
	// 引き直さずワークスペースをそのまま渡す。SuperAdmin 経路は NoWorkspace のままで、
	// usecase が req.CompanyID から解決する。
	targetWorkspace := domain.NoWorkspace()

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
		// 招待先が決まらないので、未所属の CompanyAdmin は招待できない。
		if _, affiliated := user.WorkspaceRef().WorkspaceID(); !affiliated {
			c.JSON(http.StatusForbidden, gin.H{"error": "company_admin_without_company"})
			return
		}
		// CompanyAdmin の招待先は常に自社（actor の所属ワークスペース）に固定する。
		// リクエストの companyId は無視する（他社宛の指定を握り潰す）。
		targetWorkspace = user.WorkspaceRef()
		req.CompanyID = 0
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	in := usecase.CreateAdminInvitationInput{
		CompanyID:       req.CompanyID,
		TargetWorkspace: targetWorkspace,
		Email:           req.Email,
		Role:            req.Role,
		Name:            req.Name,
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
		case errors.Is(err, usecase.ErrInvitationUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error":   "user_already_exists",
				"message": "この email のユーザーは既に存在します。パスワードの再発行は別途行ってください。",
			})
		default:
			// Cognito のスロットリング / IAM 権限不足 / DB 障害等はサーバ側障害。
			// 5xx で監視に拾わせ、AWS の内部メッセージ（pool id 等）を応答に出さない。
			log.Printf("createWithTemporaryPassword: unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		}
		return
	}
	c.JSON(http.StatusCreated, tempPasswordInvitationResponse{
		Invitation:        out.Invitation,
		TemporaryPassword: out.TemporaryPassword,
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
