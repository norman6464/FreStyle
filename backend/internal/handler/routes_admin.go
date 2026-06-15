package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/ses"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// newAuditMiddleware は管理者の変更操作（会社の有効/無効・従業員の停止/削除・招待・利用申請の承認/却下）
// を監査ログに記録する middleware を作る。記録は best-effort（監査記録の失敗で本処理は壊さない）。
func newAuditMiddleware(db *gorm.DB) gin.HandlerFunc {
	recorder := usecase.NewRecordAuditEventUseCase(persistence.NewAuditRepository(db))
	return middleware.AuditLog(func(ctx context.Context, e middleware.AuditEntry) {
		_ = recorder.Execute(ctx, usecase.RecordAuditEventInput{
			ActorID:    e.ActorID,
			ActorEmail: e.ActorEmail,
			ActorRole:  e.ActorRole,
			Action:     e.Action,
			TargetID:   e.TargetID,
		})
	})
}

// registerAdminRoutes は管理者向けの招待管理エンドポイントを登録する。
// 認可（super_admin / company_admin のみ）は handler 層で実施する。
// audit は変更操作を監査ログに記録する middleware（router で生成して共有する）。
func registerAdminRoutes(g *gin.RouterGroup, deps *routeDeps, audit gin.HandlerFunc) {
	companyRepo := persistence.NewCompanyRepository(deps.db)
	companyHandler := NewAdminCompanyHandler(
		usecase.NewListCompaniesUseCase(companyRepo),
		usecase.NewListCompanyStatsUseCase(companyRepo, persistence.NewCompanyStatsRepository(deps.db)),
		usecase.NewSetCompanyActiveUseCase(companyRepo),
	)
	g.GET("/admin/companies", companyHandler.List)
	// 会社横断ビュー（各社のメンバー集計付き。super_admin 専用）。
	g.GET("/admin/companies/stats", companyHandler.Stats)
	// 会社アカウントの有効/無効（super_admin 専用。無効化でその会社の全ユーザーを利用不可に）。
	g.PATCH("/admin/companies/:id/active", audit, companyHandler.SetActive)

	// 従業員管理（自社の従業員一覧 + 各従業員の AI 利用可否を個別上書き）。
	memberRepo := persistence.NewUserRepository(deps.db)
	memberHandler := NewAdminMemberHandler(
		usecase.NewListCompanyMembersUseCase(memberRepo),
		usecase.NewUpdateMemberAiAccessUseCase(memberRepo),
		usecase.NewSetMemberActiveUseCase(memberRepo),
		usecase.NewSoftDeleteMemberUseCase(memberRepo),
	)
	g.GET("/admin/members", memberHandler.List)
	g.PATCH("/admin/members/:userId/ai-access", memberHandler.UpdateAiAccess)
	// 従業員アカウントの有効/無効（停止）と論理削除（super_admin は全社 / company_admin は自社）。
	g.PATCH("/admin/members/:userId/active", audit, memberHandler.SetActive)
	g.DELETE("/admin/members/:userId", audit, memberHandler.Delete)

	// AdminInvitation — SES マジックリンク方式（UUID token 発行 + SES でメール送信）。
	adminInvRepo := persistence.NewAdminInvitationRepository(deps.db)

	var sender usecase.MagicLinkSender
	var linkBuilder usecase.LinkBuilder
	var mailBuilder usecase.MailBuilder

	switch {
	case deps.cfg.SES.FromAddress == "" || deps.cfg.AppBaseURL == "":
		// ローカルでは送信をスキップしてフローだけ通す（usecase 側でリンクをログに残す）。
		log.Printf("WARN: SES_FROM_ADDRESS or APP_BASE_URL not set — invitation emails will NOT be sent (token will be logged instead)")
	default:
		sesClient, err := ses.NewClient(context.Background(), deps.cfg.SES.Region, deps.cfg.SES.FromAddress)
		if err != nil {
			log.Printf("WARN: SES client init failed (invitation emails will not be sent): %v", err)
		} else {
			sender = sesClient
			baseURL := deps.cfg.AppBaseURL
			linkBuilder = func(token string) string {
				return ses.MagicLinkURL(baseURL, token)
			}
			mailBuilder = ses.BuildInvitationMail
		}
	}

	adminInvHandler := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(adminInvRepo),
		usecase.NewCreateAdminInvitationUseCase(adminInvRepo, sender, linkBuilder, mailBuilder),
		usecase.NewCancelAdminInvitationUseCase(adminInvRepo),
	)
	g.GET("/admin/invitations", adminInvHandler.List)
	g.POST("/admin/invitations", audit, adminInvHandler.Create)
	g.DELETE("/admin/invitations/:id", audit, adminInvHandler.Cancel)

	// 監査ログ閲覧（super_admin 専用）。記録は audit middleware が横断的に行う。
	auditHandler := NewAdminAuditHandler(usecase.NewListAuditEventsUseCase(persistence.NewAuditRepository(deps.db)))
	g.GET("/admin/audit-events", auditHandler.List)
}
