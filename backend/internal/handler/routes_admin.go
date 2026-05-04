package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	cognitoinfra "github.com/norman6464/FreStyle/backend/internal/infra/cognito"
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

	// Phase 25: AdminInvitation — Cognito AdminCreateUser で招待メールを送信
	adminInvRepo := repository.NewAdminInvitationRepository(deps.db)

	var cognitoClient repository.CognitoAdminClient
	if deps.cfg.Cognito.UserPoolID != "" {
		ac, err := cognitoinfra.NewAdminClient(context.Background(), "ap-northeast-1", deps.cfg.Cognito.UserPoolID)
		if err != nil {
			log.Printf("WARN: Cognito admin client init failed, falling back to stub: %v", err)
			cognitoClient = repository.NewStubCognitoAdminClient()
		} else {
			cognitoClient = ac
		}
	} else {
		log.Printf("WARN: COGNITO_USER_POOL_ID not set — invitation emails will NOT be sent (stub mode)")
		cognitoClient = repository.NewStubCognitoAdminClient()
	}

	adminInvHandler := NewAdminInvitationHandler(
		usecase.NewListAdminInvitationsUseCase(adminInvRepo),
		usecase.NewCreateAdminInvitationUseCase(adminInvRepo, cognitoClient),
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
