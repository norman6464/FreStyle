package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerChatRoutes は AI チャットセッションの REST + SSE エンドポイントを登録する。
// 会社設定で trainee の AI 利用が無効化されている場合は RequireAiChatEnabled ゲートで 403。
func registerChatRoutes(g *gin.RouterGroup, deps *routeDeps) {
	aiSessionRepo := persistence.NewAiChatSessionRepository(deps.db)
	activityRepo := persistence.NewUserDailyActivityRepository(deps.db)
	aiHandler := NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(aiSessionRepo),
		usecase.NewCreateAiChatSessionUseCase(aiSessionRepo, activityRepo),
		usecase.NewGetAiChatSessionUseCase(aiSessionRepo),
		usecase.NewUpdateAiChatSessionTitleUseCase(aiSessionRepo),
		usecase.NewDeleteAiChatSessionUseCase(aiSessionRepo),
		usecase.NewGetAiChatMessagesUseCase(deps.msgRepo),
	)

	// ai-chat 配下は会社の AI 有効化ゲートを通す（管理者・会社未所属は常に通過）。
	gate := usecase.NewAiChatEnabledForUserUseCase(persistence.NewCompanyRepository(deps.db))
	chat := g.Group("")
	chat.Use(middleware.RequireAiChatEnabled(gate))

	chat.GET("/ai-chat/sessions", aiHandler.GetSessions)
	chat.POST("/ai-chat/sessions", aiHandler.CreateSession)
	chat.GET("/ai-chat/sessions/:id", aiHandler.GetSession)
	chat.PUT("/ai-chat/sessions/:id", aiHandler.UpdateSessionTitle)
	chat.DELETE("/ai-chat/sessions/:id", aiHandler.DeleteSession)
	chat.GET("/ai-chat/sessions/:id/messages", aiHandler.GetMessages)

	// 添付用 presigned PUT URL。note image と同じバケットを ai-chat/ prefix で再利用する。
	attachmentPresigner := newAiChatAttachmentPresignerOrFallback(deps)
	attachmentHandler := NewAiChatAttachmentHandler(
		usecase.NewIssueAiChatAttachmentUploadURLUseCase(attachmentPresigner),
	)
	chat.POST("/ai-chat/attachments/upload-url", attachmentHandler.IssueUploadURL)

	// SSE ストリーミング。Bedrock / DynamoDB 初期化失敗時は nil なので登録自体をスキップする。
	if deps.bedrockClient != nil && deps.msgRepo != nil {
		downloader := newAiChatAttachmentDownloaderOrNil(deps)
		sseHandler := NewAiChatSseHandler(
			usecase.NewSendAiMessageStreamUseCase(aiSessionRepo, deps.msgRepo, deps.bedrockClient, downloader, activityRepo),
		)
		chat.POST("/ai-chat/stream", sseHandler.Handle)
	}
}

// newAiChatAttachmentPresignerOrFallback は本番では S3 presigner、dev / 未設定時は stub を返す（fail open）。
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

// newAiChatAttachmentDownloaderOrNil は本番では S3 downloader を返す。
// bucket 未設定 / 失敗時は nil を返し、SSE usecase 側で添付なしと同じ振る舞いにフォールバックする。
func newAiChatAttachmentDownloaderOrNil(deps *routeDeps) usecase.AttachmentDownloader {
	bucket := deps.cfg.S3.NoteImagesBucket
	if bucket == "" {
		return nil
	}
	// ai-chat/ prefix のみ読み出し許可（他用途の読み出しを防止）。
	dl, err := infraS3.NewDownloader(context.Background(), deps.cfg.S3.Region, bucket, "ai-chat/")
	if err != nil {
		log.Printf("[ai-chat-attachment] failed to init S3 downloader (%v) — attachments will be ignored", err)
		return nil
	}
	return dl
}
