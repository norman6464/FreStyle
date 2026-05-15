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

// registerChatRoutes は AI チャットセッションの REST + SSE エンドポイントを登録する。
// WebSocket は registerWebSocketRoutes 側で別途登録する（SSE 移行完了まで並行運用）。
func registerChatRoutes(g *gin.RouterGroup, deps *routeDeps) {
	aiSessionRepo := persistence.NewAiChatSessionRepository(deps.db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo),
		usecase.NewGetAiChatSessionUseCase(aiSessionRepo),
		usecase.NewUpdateAiChatSessionTitleUseCase(aiSessionRepo),
		usecase.NewDeleteAiChatSessionUseCase(aiSessionRepo),
		usecase.NewGetAiChatMessagesUseCase(deps.msgRepo),
	)
	g.GET("/ai-chat/sessions", aiHandler.GetSessions)
	g.POST("/ai-chat/sessions", aiHandler.CreateSession)
	g.GET("/ai-chat/sessions/:id", aiHandler.GetSession)
	g.PUT("/ai-chat/sessions/:id", aiHandler.UpdateSessionTitle)
	g.DELETE("/ai-chat/sessions/:id", aiHandler.DeleteSession)
	g.GET("/ai-chat/sessions/:id/messages", aiHandler.GetMessages)

	// 添付ファイル用 presigned PUT URL（PR-G1: 画像）。
	// note image 系と同じく NOTE_IMAGES_BUCKET（= note-images バケット）を再利用し、
	// `ai-chat/{userId}/{uuid}.{ext}` の prefix で運用する。bucket policy / IAM は
	// 既存ノート画像と共有（runtime/iam.yml の `${NoteImagesBucketArn}/*` でカバー）。
	attachmentPresigner := newAiChatAttachmentPresignerOrFallback(deps)
	attachmentHandler := NewAiChatAttachmentHandler(
		usecase.NewIssueAiChatAttachmentUploadURLUseCase(attachmentPresigner),
	)
	g.POST("/ai-chat/attachments/upload-url", attachmentHandler.IssueUploadURL)

	// SSE ストリーミング（汎用 AI チャットの token 単位送信）。
	// Bedrock / DynamoDB の初期化に失敗していると bedrockClient / msgRepo は nil。
	// nil sentinel は handler 側で 503 にする。
	if deps.bedrockClient != nil && deps.msgRepo != nil {
		downloader := newAiChatAttachmentDownloaderOrNil(deps)
		sseHandler := NewAiChatSseHandler(
			usecase.NewSendAiMessageStreamUseCase(aiSessionRepo, deps.msgRepo, deps.bedrockClient, downloader),
		)
		g.POST("/ai-chat/stream", sseHandler.Handle)
	}
}

// newAiChatAttachmentPresignerOrFallback は本番経路では infra/s3.Presigner を、
// dev / NOTE_IMAGES_BUCKET 未設定時は stub presigner を返す。Note 系と同じ fail open。
func newAiChatAttachmentPresignerOrFallback(deps *routeDeps) repository.AiChatAttachmentPresigner {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		log.Printf("[ai-chat-attachment] NOTE_IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubAiChatAttachmentPresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Printf("[ai-chat-attachment] failed to init S3 presigner (%v) — falling back to stub", err)
		return persistence.NewStubAiChatAttachmentPresigner(bucket)
	}
	return persistence.NewAiChatAttachmentPresigner(pre)
}

// newAiChatAttachmentDownloaderOrNil は本番では S3 GetObject downloader を返す。
// bucket 未設定 / 初期化失敗時は nil を返し、SSE usecase 側で「添付なしと同じ振る舞い」にフォールバック。
func newAiChatAttachmentDownloaderOrNil(deps *routeDeps) usecase.AttachmentDownloader {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		return nil
	}
	// `ai-chat/` prefix のみ読み出し許可（他用途の S3 オブジェクト読み出しを防止）。
	dl, err := infraS3.NewDownloader(context.Background(), deps.cfg.S3.Region, bucket, "ai-chat/")
	if err != nil {
		log.Printf("[ai-chat-attachment] failed to init S3 downloader (%v) — attachments will be ignored", err)
		return nil
	}
	return dl
}
