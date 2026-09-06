package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/usecase/profile"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerProfileRoutes は profile 関連の REST エンドポイントを登録する。
func registerProfileRoutes(g *gin.RouterGroup, deps *routeDeps) {
	profileRepo := persistence.NewProfileRepository(deps.db)
	profileHandler := NewProfileHandler(
		profile.NewGetProfileUseCase(profileRepo),
		profile.NewUpdateProfileUseCase(profileRepo),
		deps.userRepo,
	)
	// :userId は数字 / "me" の両方を受ける。/update はフロント互換の別 path。
	g.GET("/profile/:userId", profileHandler.Get)
	g.PUT("/profile/:userId", profileHandler.Update)
	g.PUT("/profile/:userId/update", profileHandler.Update) //apispec:allow フロント互換の別 path（正規は PUT /profile/:userId）

	// Profile アイコン画像の S3 presigned-url（リッチテキスト画像と同じバケットを profiles/ prefix で共有）。
	profileImageHandler := NewProfileImageHandler(
		profile.NewIssueProfileImageUploadURLUseCase(
			newProfileImagePresignerOrFallback(deps),
		),
	)
	g.POST("/profile/:userId/image/presigned-url", profileImageHandler.IssueUploadURL)
}

func newProfileImagePresignerOrFallback(deps *routeDeps) repository.ProfileImagePresigner {
	bucket := deps.cfg.S3.ImagesBucket
	if bucket == "" {
		log.Printf("[profile] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubProfileImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[profile] failed to init S3 presigner (%v) — falling back to stub", err)
		return persistence.NewStubProfileImagePresigner(bucket)
	}
	return persistence.NewProfileImagePresigner(pre)
}
