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

// newRichTextImagePresignerOrFallback は IMAGES_BUCKET 未設定なら stub にフォールバックする
// （明示的にローカル開発用と分かる状態なので安全）。bucket が設定されているのに
// infraS3.NewPresigner が失敗する場合は fallback しない（CodeRabbit 指摘・段1b で発見。
// kb_page_handler 側の newKbImagePresignerOrFallback の doc も参照）— 黙って stub
// （未署名 URL）へ倒すと呼び出し元は 200 を返し続け、クライアントは成功と誤認したまま
// S3 PUT だけが失敗するため、起動を失敗させる。
func newRichTextImagePresignerOrFallback(deps *routeDeps) repository.RichTextImagePresigner {
	bucket := deps.cfg.S3.ImagesBucket
	if bucket == "" {
		log.Printf("[rich-text-image] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubRichTextImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Fatalf("[rich-text-image] IMAGES_BUCKET=%q is set but S3 presigner init failed: %v", bucket, err)
	}
	return persistence.NewRichTextImagePresigner(pre)
}
