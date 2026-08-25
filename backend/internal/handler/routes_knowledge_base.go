package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerKnowledgeBaseRoutes はナレッジ基盤のページ操作のエンドポイントを登録する。
//
// ワークスペースは URL の slug から middleware が解決するので、ルートはすべて
// /kb/workspaces/:workspaceSlug 以下に置き、その middleware を通す group に登録する
// （通し忘れたルートはテナント未確定のまま handler に入るため、group をここ 1 箇所に閉じる）。
func registerKnowledgeBaseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	// ナレッジ基盤は GORM を通さない（スキーマの正本が明示 SQL）ため、
	// GORM が持つ接続プールから *sql.DB を借りて sqlc 実装へ渡す。
	sqlDB, err := deps.db.DB()
	if err != nil {
		log.Printf("WARN: *sql.DB を取得できないためナレッジ基盤の API を登録しません: %v", err)
		return
	}
	registerKnowledgeBaseRoutesWith(
		g,
		persistence.NewKnowledgeBaseRepository(sqlDB),
		persistence.NewKnowledgeBasePermissionRepository(sqlDB),
	)
}

// registerKnowledgeBaseRoutesWith は repository を受け取ってルートと middleware を組み立てる。
// 本番の wiring とテストが同じ 1 箇所を通るようにするために切り出してある
// （テストがルート表を書き写すと、本番だけ middleware が抜けた配線ミスを見逃す）。
func registerKnowledgeBaseRoutesWith(
	g *gin.RouterGroup,
	pages repository.KnowledgeBaseRepository,
	permissions repository.KnowledgeBasePermissionRepository,
) {
	h := NewKnowledgeBasePageHandler(
		usecase.NewCheckPagePermissionUseCase(permissions),
		usecase.NewCanEditPageSubtreeUseCase(permissions),
		usecase.NewListViewablePagesUseCase(permissions),
		usecase.NewGetPageUseCase(pages),
		usecase.NewCreatePageUseCase(pages),
		usecase.NewRenamePageUseCase(pages),
		usecase.NewMovePageUseCase(pages),
		usecase.NewArchivePageUseCase(pages),
		usecase.NewUnarchivePageUseCase(pages),
		usecase.NewReplacePageBlocksUseCase(pages),
	)

	kb := g.Group("", middleware.KnowledgeBaseWorkspace(
		usecase.NewResolveWorkspaceUseCase(pages, permissions),
	))
	kb.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Tree)
	kb.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Create)
	kb.GET("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Get)
	kb.PATCH("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Rename)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/move", h.Move)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/archive", h.Archive)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/unarchive", h.Unarchive)
	kb.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/content", h.ReplaceContent)
}
