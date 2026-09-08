package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	infraGCS "github.com/norman6464/FreStyle/backend/internal/infra/gcs"
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

	// Profile アイコン画像の presigned-url（リッチテキスト画像と同じバケットを profiles/ prefix で共有）。
	profileImageHandler := NewProfileImageHandler(
		profile.NewIssueProfileImageUploadURLUseCase(
			newProfileImagePresignerOrFallback(deps),
		),
	)
	g.POST("/profile/:userId/image/presigned-url", profileImageHandler.IssueUploadURL)
}

// newProfileImagePresignerOrFallback は IMAGES_BUCKET 未設定なら stub にフォールバックする
// （明示的にローカル開発用と分かる状態なので安全）。bucket が設定されているのに
// infraGCS.NewPresigner が失敗する場合は fallback しない（CodeRabbit 指摘・段1b で発見。
// kb_page_handler 側の newKbImagePresignerOrFallback の doc も参照）。
func newProfileImagePresignerOrFallback(deps *routeDeps) repository.ProfileImagePresigner {
	bucket := deps.cfg.Images.Bucket
	if bucket == "" {
		log.Printf("[profile] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubProfileImagePresigner("stub-bucket")
	}
	pre, err := infraGCS.NewPresigner(context.Background(), bucket)
	if err != nil {
		log.Fatalf("[profile] IMAGES_BUCKET=%q is set but GCS presigner init failed: %v", bucket, err)
	}
	return persistence.NewProfileImagePresigner(pre)
}
