package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerNoteRoutes は Note CRUD・Note 画像 presigned URL・SessionNote のエンドポイントを登録する。
func registerNoteRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 10: Note CRUD
	noteRepo := repository.NewNoteRepository(deps.db)
	noteHandler := NewNoteHandler(
		usecase.NewListNotesByUserIDUseCase(noteRepo),
		usecase.NewCreateNoteUseCase(noteRepo),
		usecase.NewUpdateNoteUseCase(noteRepo),
		usecase.NewDeleteNoteUseCase(noteRepo),
	)
	g.GET("/notes", noteHandler.List)
	g.POST("/notes", noteHandler.Create)
	g.PUT("/notes/:id", noteHandler.Update)
	g.DELETE("/notes/:id", noteHandler.Delete)

	// Phase 11: Note image (S3 presigned upload)
	noteImageHandler := NewNoteImageHandler(
		usecase.NewIssueNoteImageUploadURLUseCase(newNoteImagePresignerOrFallback(deps)),
	)
	g.POST("/notes/images/upload-url", noteImageHandler.IssueUploadURL)

	// Phase 12: SessionNote (セッション固有ノート)
	sessionNoteRepo := repository.NewSessionNoteRepository(deps.db)
	sessionNoteHandler := NewSessionNoteHandler(
		usecase.NewGetSessionNoteUseCase(sessionNoteRepo),
		usecase.NewUpsertSessionNoteUseCase(sessionNoteRepo),
	)
	g.GET("/sessions/:sessionId/note", sessionNoteHandler.Get)
	g.PUT("/sessions/:sessionId/note", sessionNoteHandler.Upsert)
}

// newNoteImagePresignerOrFallback は本番では infra/s3.Presigner で real な PUT presign を返し、
// NOTE_IMAGES_BUCKET 未設定 (ローカル / dev) の場合だけ stub にフォールバックする。
// presigner 失敗時も fail open で stub に降格 (Note image 機能だけ利用不可になる)。
func newNoteImagePresignerOrFallback(deps *routeDeps) repository.NoteImagePresigner {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		log.Printf("[note-image] NOTE_IMAGES_BUCKET unset — using stub presigner (DEV)")
		return repository.NewStubNoteImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[note-image] failed to init S3 presigner (%v) — falling back to stub", err)
		return repository.NewStubNoteImagePresigner(bucket)
	}
	return repository.NewNoteImagePresigner(pre)
}
