package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerCompanySettingsRoutes は会社設定（trainee への AI 有効化）の取得・更新を登録する。
func registerCompanySettingsRoutes(g *gin.RouterGroup, deps *routeDeps) {
	companies := persistence.NewCompanyRepository(deps.db)
	h := NewCompanySettingsHandler(
		usecase.NewGetCompanyAiChatSettingUseCase(companies),
		usecase.NewUpdateCompanyAiChatSettingUseCase(companies),
	)
	g.GET("/company/settings", h.Get)
	g.PUT("/company/settings", h.Update)
}
