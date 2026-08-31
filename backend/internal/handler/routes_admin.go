package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/cognito"
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
	// 1 箇所書き忘れただけで穴になるため多層防御を敷く（FRESTYLE-76 / FRESTYLE-228）。
	g := parent.Group("", middleware.RequireAdmin())

	// 従業員管理（自社の従業員一覧・有効/無効・論理削除）。
	memberRepo := persistence.NewUserRepository(deps.db)
	memberHandler := NewAdminMemberHandler(
		usecase.NewListCompanyMembersUseCase(memberRepo),
		usecase.NewSetMemberActiveUseCase(memberRepo),
		usecase.NewSoftDeleteMemberUseCase(memberRepo),
		usecase.NewGetCompanyLearningSummaryUseCase(persistence.NewCompanyLearningActivityRepository(deps.db)),
	)
	g.GET("/admin/members", memberHandler.List)
	// company_admin のホーム用: 自社メンバーの学習状況サマリー(FRESTYLE-103)。
	g.GET("/admin/members/learning-summary", memberHandler.LearningSummary)
	// 従業員アカウントの有効/無効（停止）と論理削除（super_admin は全社 / company_admin は自社）。
	g.PATCH("/admin/members/:userId/active", memberHandler.SetActive)
	g.DELETE("/admin/members/:userId", memberHandler.Delete)

	// AdminInvitation — SES マジックリンク方式（UUID token 発行 + SES でメール送信）。
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

	// 初期パスワード方式（FRESTYLE-313）。COGNITO_USER_POOL_ID 未設定なら nil を渡し、
	// handler 側で「未構成」として 400 にする（マジックリンク方式には影響しない）。
	var tempCreate *usecase.CreateTemporaryPasswordInvitationUseCase
	if deps.cfg.Cognito.UserPoolID != "" {
		if creator, err := cognito.NewAdminUserCreator(context.Background(), deps.cfg.Cognito.Region, deps.cfg.Cognito.UserPoolID); err != nil {
			log.Printf("WARN: admin user creator init failed (temporary password invitations disabled): %v", err)
		} else {
			tempCreate = usecase.NewCreateTemporaryPasswordInvitationUseCase(adminInvRepo, creator)
		}
	}

	adminInvHandler := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(adminInvRepo),
		usecase.NewCreateAdminInvitationUseCase(adminInvRepo, sender, linkBuilder, mailBuilder),
		tempCreate,
		usecase.NewCancelAdminInvitationUseCase(adminInvRepo),
	)
	g.GET("/admin/invitations", adminInvHandler.List)
	// 招待作成は Cognito ユーザー作成 / 存在確認オラクルを伴うため、管理者権限でも流量を絞る
	// （email の総当たり列挙・大量アカウント生成の抑止・FRESTYLE-313）。
	g.POST("/admin/invitations", middleware.RateLimitPerMinute(20, 10), adminInvHandler.Create)
	g.DELETE("/admin/invitations/:id", adminInvHandler.Cancel)
}
