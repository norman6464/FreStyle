package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/infra/ses"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAdminRoutes は管理者向けの招待管理エンドポイントを登録する。
// 認可（super_admin / company_admin のみ）は handler 層で実施する。
func registerAdminRoutes(g *gin.RouterGroup, deps *routeDeps) {
	companyHandler := NewAdminCompanyHandler(
		usecase.NewListCompaniesUseCase(persistence.NewCompanyRepository(deps.db)),
	)
	g.GET("/admin/companies", companyHandler.List)

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
	g.POST("/admin/invitations", adminInvHandler.Create)
	g.DELETE("/admin/invitations/:id", adminInvHandler.Cancel)
}
