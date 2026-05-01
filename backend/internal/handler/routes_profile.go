package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
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
	// （prefix `profiles/` で分離）。
	profileImageHandler := NewProfileImageHandler(
		usecase.NewIssueProfileImageUploadURLUseCase(
			newProfileImagePresignerOrFallback(deps),
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

// newProfileImagePresignerOrFallback は本番では infra/s3.Presigner で real な PUT presign を返し、
// 環境変数 NOTE_IMAGES_BUCKET が未設定 (ローカル / dev) の場合だけ stub にフォールバックする。
//
// 配信 URL のベース (CDN) は config.S3.NoteImagesCDNBase を最優先で使う。
// 未指定なら virtual-hosted-style (https://<bucket>.s3.<region>.amazonaws.com) で組み立てる。
func newProfileImagePresignerOrFallback(deps *routeDeps) repository.ProfileImagePresigner {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		log.Printf("[profile] NOTE_IMAGES_BUCKET unset — using stub presigner (DEV)")
		return repository.NewStubProfileImagePresigner("stub-bucket", deps.cfg.S3.NoteImagesCDNBase)
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[profile] failed to init S3 presigner (%v) — falling back to stub", err)
		return repository.NewStubProfileImagePresigner(bucket, deps.cfg.S3.NoteImagesCDNBase)
	}
	cdnBase := deps.cfg.S3.NoteImagesCDNBase
	if cdnBase == "" {
		cdnBase = "https://" + bucket + ".s3." + deps.cfg.S3.Region + ".amazonaws.com"
	}
	return repository.NewProfileImagePresigner(pre, cdnBase)
}
