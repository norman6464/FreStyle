package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerProfileRoutes は profile / user-stats 関連の REST エンドポイントを登録する。
func registerProfileRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 5: プロフィール
	profileRepo := repository.NewProfileRepository(deps.db)
	profileHandler := NewProfileHandler(
		usecase.NewGetProfileUseCase(profileRepo),
		usecase.NewUpdateProfileUseCase(profileRepo),
		deps.userRepo,
	)
	// /profile/:userId は数字 / "me" の両方を受ける（handler.resolveUserID で current user 解決）。
	// フロント互換のため /profile/:userId/update PUT / /profile/:userId/image/presigned-url POST も提供する。
	g.GET("/profile/:userId", profileHandler.Get)
	g.PUT("/profile/:userId", profileHandler.Update)
	g.PUT("/profile/:userId/update", profileHandler.Update)

	// Profile アイコン画像の S3 presigned-url。bucket / CDN は note image と同じインフラを共有する
	// （別 prefix `profiles/` で分離）。AWS SDK 統合は別 issue。
	profileImageHandler := NewProfileImageHandler(
		usecase.NewIssueProfileImageUploadURLUseCase(
			repository.NewStubProfileImagePresigner("frestyle-prod-note-images", "https://normanblog.com"),
		),
	)
	g.POST("/profile/:userId/image/presigned-url", profileImageHandler.IssueUploadURL)

	// Phase 6: ユーザー統計
	statsHandler := NewUserStatsHandler(
		usecase.NewGetUserStatsUseCase(repository.NewUserStatsRepository(deps.db)),
	)
	// 旧 path /user-stats/:userId と Spring 風 /users/:userId/stats の両方を提供する。
	g.GET("/user-stats/:userId", statsHandler.Get)
	g.GET("/users/:userId/stats", statsHandler.Get)
}
