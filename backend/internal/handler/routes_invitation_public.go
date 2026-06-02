package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerInvitationPublicRoutes は招待 token 検証の認証不要エンドポイントを登録する。
// 無効・期限切れは 404 を返し、招待の存在有無を漏らさない。
func registerInvitationPublicRoutes(g *gin.RouterGroup, deps *routeDeps) {
	invRepo := persistence.NewAdminInvitationRepository(deps.db)
	companies := persistence.NewCompanyRepository(deps.db)
	h := NewPublicInvitationHandler(usecase.NewValidateInvitationTokenUseCase(invRepo, companies))
	// token 総当たり（招待 token の列挙）を緩和するため per-IP レートリミットを掛ける。
	g.GET("/invitations/accept/:token", middleware.RateLimitPerMinute(30, 10), h.Validate)
}
