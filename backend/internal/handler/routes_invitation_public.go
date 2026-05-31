package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerInvitationPublicRoutes は招待 token 検証の認証不要エンドポイントを登録する。
// 無効・期限切れは 404 を返し、招待の存在有無を漏らさない。
func registerInvitationPublicRoutes(g *gin.RouterGroup, deps *routeDeps) {
	invRepo := persistence.NewAdminInvitationRepository(deps.db)
	companies := persistence.NewCompanyRepository(deps.db)
	h := NewPublicInvitationHandler(usecase.NewValidateInvitationTokenUseCase(invRepo, companies))
	g.GET("/invitations/accept/:token", h.Validate)
}
