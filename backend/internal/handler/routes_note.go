package handler

import (
	"github.com/gin-gonic/gin"
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
		usecase.NewIssueNoteImageUploadURLUseCase(
			repository.NewStubNoteImagePresigner("frestyle-prod-note-images"),
		),
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
