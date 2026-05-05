package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/infra/ses"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAdminRoutes は管理者向けの招待管理・シナリオ管理エンドポイントを登録する。
// 認可（super_admin / company_admin のみ操作可能）は handler 層で実施する。
func registerAdminRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Company 一覧 (SuperAdmin 専用)
	companyHandler := NewAdminCompanyHandler(
		usecase.NewListCompaniesUseCase(repository.NewCompanyRepository(deps.db)),
	)
	g.GET("/admin/companies", companyHandler.List)

	// AdminInvitation — SES マジックリンク方式（Path B）。
	// Cognito 事前作成は撤去し、ここでは UUID token を発行 + SES で独自メールを送る。
	adminInvRepo := repository.NewAdminInvitationRepository(deps.db)

	var sender usecase.MagicLinkSender
	var linkBuilder usecase.LinkBuilder
	var mailBuilder usecase.MailBuilder

	switch {
	case deps.cfg.SES.FromAddress == "" || deps.cfg.AppBaseURL == "":
		// SES 未設定 / APP_BASE_URL 未設定のローカル環境では送信をスキップしてフローを通すだけにする。
		// 本番では必ず両方設定する前提（usecase 側でログにリンクを残す）。
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

	// Phase 26: AdminScenario
	adminScenarioRepo := repository.NewAdminScenarioRepository(deps.db)
	adminScenarioHandler := NewAdminScenarioHandler(
		usecase.NewListAdminScenariosUseCase(adminScenarioRepo),
		usecase.NewCreateAdminScenarioUseCase(adminScenarioRepo),
		usecase.NewUpdateAdminScenarioUseCase(adminScenarioRepo),
		usecase.NewDeleteAdminScenarioUseCase(adminScenarioRepo),
	)
	g.GET("/admin/scenarios", adminScenarioHandler.List)
	g.POST("/admin/scenarios", adminScenarioHandler.Create)
	g.PUT("/admin/scenarios/:id", adminScenarioHandler.Update)
	g.DELETE("/admin/scenarios/:id", adminScenarioHandler.Delete)
}
