package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerInvitationPublicRoutes は招待マジックリンク受諾画面用の認証不要エンドポイントを登録する。
//
//	GET /api/v2/invitations/accept/:token  招待 token の検証 + 受諾画面用の最低限の情報返却
//
// 認可は無し（受諾前にユーザーが踏むため）。token が無効・期限切れの場合は 404 を返し、
// 「招待が存在するかどうか」のメタ情報を漏らさない。
func registerInvitationPublicRoutes(g *gin.RouterGroup, deps *routeDeps) {
	invRepo := persistence.NewAdminInvitationRepository(deps.db)
	companies := persistence.NewCompanyRepository(deps.db)
	h := NewPublicInvitationHandler(usecase.NewValidateInvitationTokenUseCase(invRepo, companies))
	g.GET("/invitations/accept/:token", h.Validate)
}
