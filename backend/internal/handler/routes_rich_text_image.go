package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/richtextimage"
)

// registerRichTextImageRoutes はリッチテキスト画像 presigned URL のエンドポイントを登録する。
func registerRichTextImageRoutes(g *gin.RouterGroup, deps *routeDeps) {
	richTextImageHandler := NewRichTextImageHandler(
		richtextimage.NewIssueRichTextImageUploadURLUseCase(newRichTextImagePresignerOrFallback(deps)),
	)
	g.POST("/rich-text/images/upload-url", richTextImageHandler.IssueUploadURL)
}

// newRichTextImagePresignerOrFallback は本番では real な presigner、IMAGES_BUCKET 未設定や
// 初期化失敗時は stub にフォールバックする（fail open）。
func newRichTextImagePresignerOrFallback(deps *routeDeps) repository.RichTextImagePresigner {
	bucket := deps.cfg.S3.ImagesBucket
	if bucket == "" {
		log.Printf("[rich-text-image] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubRichTextImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[rich-text-image] failed to init S3 presigner (%v) — falling back to stub", err)
		return persistence.NewStubRichTextImagePresigner(bucket)
	}
	return persistence.NewRichTextImagePresigner(pre)
}
