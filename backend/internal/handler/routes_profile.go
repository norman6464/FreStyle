package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerProfileRoutes は profile / user-stats 関連の REST エンドポイントを登録する。
func registerProfileRoutes(g *gin.RouterGroup, deps *routeDeps) {
	profileRepo := persistence.NewProfileRepository(deps.db)
	profileHandler := NewProfileHandler(
		usecase.NewGetProfileUseCase(profileRepo),
		usecase.NewUpdateProfileUseCase(profileRepo),
		deps.userRepo,
	)
	// :userId は数字 / "me" の両方を受ける。/update はフロント互換の別 path。
	g.GET("/profile/:userId", profileHandler.Get)
	g.PUT("/profile/:userId", profileHandler.Update)
	g.PUT("/profile/:userId/update", profileHandler.Update) //apispec:allow フロント互換の別 path（正規は PUT /profile/:userId）

	// Profile アイコン画像の S3 presigned-url（note image と同じバケットを profiles/ prefix で共有）。
	profileImageHandler := NewProfileImageHandler(
		usecase.NewIssueProfileImageUploadURLUseCase(
			newProfileImagePresignerOrFallback(deps),
		),
	)
	g.POST("/profile/:userId/image/presigned-url", profileImageHandler.IssueUploadURL)

	// ユーザー統計。2 つの path は互換のため両方提供する。
	statsHandler := NewUserStatsHandler(
		usecase.NewGetUserStatsUseCase(persistence.NewUserStatsRepository(deps.db)),
	)
	g.GET("/user-stats/:userId", statsHandler.Get) //apispec:allow 互換用エイリアス（正規は GET /users/:userId/stats）
	g.GET("/users/:userId/stats", statsHandler.Get)

	// オンボーディング完了（自分自身の users.onboarded_at を更新）。
	onboardingHandler := NewOnboardingHandler(usecase.NewCompleteOnboardingUseCase(deps.userRepo))
	g.POST("/profile/me/onboarding/complete", onboardingHandler.Complete)
}

// newProfileImagePresignerOrFallback は本番では real な presigner、NOTE_IMAGES_BUCKET 未設定や
// 初期化失敗時は stub にフォールバックする。CDN ベースは未指定時 virtual-hosted-style で組み立てる。
func newProfileImagePresignerOrFallback(deps *routeDeps) repository.ProfileImagePresigner {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		log.Printf("[profile] NOTE_IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubProfileImagePresigner("stub-bucket", deps.cfg.S3.NoteImagesCDNBase)
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[profile] failed to init S3 presigner (%v) — falling back to stub", err)
		return persistence.NewStubProfileImagePresigner(bucket, deps.cfg.S3.NoteImagesCDNBase)
	}
	cdnBase := deps.cfg.S3.NoteImagesCDNBase
	if cdnBase == "" {
		cdnBase = "https://" + bucket + ".s3." + deps.cfg.S3.Region + ".amazonaws.com"
	}
	return persistence.NewProfileImagePresigner(pre, cdnBase)
}
