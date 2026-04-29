package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerPhraseRoutes は会話テンプレートとお気に入りフレーズの REST エンドポイントを登録する。
func registerPhraseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// Phase 18: ConversationTemplate
	templateHandler := NewConversationTemplateHandler(
		usecase.NewListConversationTemplatesUseCase(repository.NewConversationTemplateRepository(deps.db)),
	)
	g.GET("/conversation-templates", templateHandler.List)

	// Phase 19: FavoritePhrase
	favRepo := repository.NewFavoritePhraseRepository(deps.db)
	favHandler := NewFavoritePhraseHandler(
		usecase.NewListFavoritePhrasesUseCase(favRepo),
		usecase.NewAddFavoritePhraseUseCase(favRepo),
		usecase.NewDeleteFavoritePhraseUseCase(favRepo),
	)
	g.GET("/favorite-phrases", favHandler.List)
	g.POST("/favorite-phrases", favHandler.Add)
	g.DELETE("/favorite-phrases/:id", favHandler.Remove)
}
