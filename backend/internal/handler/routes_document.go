package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerDocumentRoutes はリッチテキスト文書 CRUD のエンドポイントを登録する。
func registerDocumentRoutes(g *gin.RouterGroup, deps *routeDeps) {
	repo := persistence.NewRichDocumentRepository(deps.db)
	h := NewDocumentHandler(
		usecase.NewGetRichDocumentUseCase(repo),
		usecase.NewCreateRichDocumentUseCase(repo),
		usecase.NewUpdateRichDocumentUseCase(repo),
		usecase.NewDeleteRichDocumentUseCase(repo),
	)
	g.POST("/documents", h.Create)
	g.GET("/documents/:id", h.Get)
	g.PUT("/documents/:id", h.Update)
	g.DELETE("/documents/:id", h.Delete)
}
