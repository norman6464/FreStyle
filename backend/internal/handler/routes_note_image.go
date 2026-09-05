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

// registerNoteImageRoutes は Note 画像 presigned URL のエンドポイントを登録する。
func registerNoteImageRoutes(g *gin.RouterGroup, deps *routeDeps) {
	noteImageHandler := NewNoteImageHandler(
		usecase.NewIssueNoteImageUploadURLUseCase(newNoteImagePresignerOrFallback(deps)),
	)
	g.POST("/notes/images/upload-url", noteImageHandler.IssueUploadURL)
}

// newNoteImagePresignerOrFallback は本番では real な presigner、NOTE_IMAGES_BUCKET 未設定や
// 初期化失敗時は stub にフォールバックする（fail open）。
func newNoteImagePresignerOrFallback(deps *routeDeps) repository.NoteImagePresigner {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		log.Printf("[note-image] NOTE_IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubNoteImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[note-image] failed to init S3 presigner (%v) — falling back to stub", err)
		return persistence.NewStubNoteImagePresigner(bucket)
	}
	return persistence.NewNoteImagePresigner(pre)
}
