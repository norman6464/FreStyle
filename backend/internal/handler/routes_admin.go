package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/ses"
	"github.com/norman6464/FreStyle/backend/internal/infra/smtp"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerAdminRoutes は管理者向けエンドポイント（会員 / 招待）を登録する。
//
// 認可は 2 段構え:
//   - 入口: middleware.RequireAdmin が非管理者（trainee 等）を落とす
//   - 詳細: どのワークスペースまで見えるかは各 handler / usecase が判定する
func registerAdminRoutes(parent *gin.RouterGroup, deps *routeDeps) {
	// 非管理者(trainee 等)は入口で落とす。個々の handler の role 検査は残すが、
	// 1 箇所書き忘れただけで穴になるため多層防御を敷く。
	g := parent.Group("", middleware.RequireAdmin())

	memberRepo := persistence.NewUserRepository(deps.db)
	memberHandler := NewAdminMemberHandler(
		usecase.NewListCompanyMembersUseCase(memberRepo),
		usecase.NewSetMemberActiveUseCase(memberRepo),
		usecase.NewSoftDeleteMemberUseCase(memberRepo),
	)
	g.GET("/admin/members", memberHandler.List)
	// 従業員アカウントの有効/無効（停止）と論理削除（super_admin は全社 / company_admin は自社）。
	g.PATCH("/admin/members/:userId/active", memberHandler.SetActive)
	g.DELETE("/admin/members/:userId", memberHandler.Delete)

	adminInvRepo := persistence.NewAdminInvitationRepository(deps.db)

	var sender usecase.MagicLinkSender
	var linkBuilder usecase.LinkBuilder
	var mailBuilder usecase.MailBuilder

	switch {
	case deps.cfg.SMTP.Host != "" && deps.cfg.SMTP.FromAddress != "" && deps.cfg.AppBaseURL != "":
		// staging: SES は使わず、box 上のメールキャッチャーへ SMTP で送る。
		// 実メールは外部へ飛ばず、受信分は mail サブドメインの Web UI で閲覧する。
		sender = smtp.NewSender(deps.cfg.SMTP.Host, deps.cfg.SMTP.Port, deps.cfg.SMTP.FromAddress)
		baseURL := deps.cfg.AppBaseURL
		linkBuilder = func(token string) string {
			return ses.MagicLinkURL(baseURL, token)
		}
		mailBuilder = ses.BuildInvitationMail
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
	// 招待作成は「その email が既に居るか」を応答の差から確かめられる面なので、
	// 管理者権限でも流量を絞る（email の総当たり列挙・大量招待の抑止）。
	g.POST("/admin/invitations", middleware.RateLimitPerMinute(20, 10), adminInvHandler.Create)
	g.DELETE("/admin/invitations/:id", adminInvHandler.Cancel)
}
