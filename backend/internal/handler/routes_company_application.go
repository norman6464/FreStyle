package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// newCompanyApplicationHandler は申請の usecase 一式を組み立てて handler を返す。
func newCompanyApplicationHandler(deps *routeDeps) *CompanyApplicationHandler {
	appRepo := persistence.NewCompanyApplicationRepository(deps.db)
	notifRepo := persistence.NewNotificationRepository(deps.db)
	return NewCompanyApplicationHandler(
		usecase.NewCreateCompanyApplicationUseCase(appRepo, deps.userRepo, notifRepo),
		usecase.NewListCompanyApplicationsUseCase(appRepo),
		usecase.NewUpdateCompanyApplicationStatusUseCase(appRepo),
	)
}

// registerCompanyApplicationPublicRoutes は認証不要の企業申請作成を登録する。
// 公開フォームのスパム対策として per-IP レートリミット（5 回/分、burst 5）を掛ける。
func registerCompanyApplicationPublicRoutes(g *gin.RouterGroup, h *CompanyApplicationHandler) {
	g.POST("/company-applications", middleware.RateLimitPerMinute(5, 5), h.Create)
}

// registerCompanyApplicationAdminRoutes は super_admin 用の一覧 / status 更新を登録する（認可は handler 層）。
func registerCompanyApplicationAdminRoutes(g *gin.RouterGroup, h *CompanyApplicationHandler) {
	g.GET("/admin/company-applications", h.List)
	g.PATCH("/admin/company-applications/:id/status", h.UpdateStatus)
}
