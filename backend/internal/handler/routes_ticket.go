package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
)

// registerTicketRoutes はチケット（設計 Ⅵ・段 1: 骨格）のエンドポイントを登録する。
//
// チケットは既存の spaces に属する（設計 Ⅱ）ので、URL は kb と同じ
// /kb/workspaces/:workspaceSlug 以下に置き、同じ middleware.KnowledgeBaseWorkspace を通す
// （registerKnowledgeBaseRoutesWith の doc と同じ理由 — group をここ 1 箇所に閉じる）。
// kb 側の routes_knowledge_base.go には触れず、別ファイルとして独立させてある
// （usecase/ticket は usecase/kb を import しない境界だが、handler 層は両方に依存してよい）。
func registerTicketRoutes(g *gin.RouterGroup, deps *routeDeps) {
	registerTicketRoutesWith(
		g,
		persistence.NewTicketRepository(deps.db),
		persistence.NewKnowledgeBasePermissionRepository(deps.db),
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewUserRepository(deps.db),
		persistence.NewTxManager(deps.db),
	)
}

// registerTicketRoutesWith は repository を受け取ってルートと middleware を組み立てる
// （本番の wiring とテストが同じ 1 箇所を通る。registerKnowledgeBaseRoutesWith と同じ理由）。
func registerTicketRoutesWith(
	g *gin.RouterGroup,
	tickets repository.TicketRepository,
	permissions repository.KnowledgeBasePermissionRepository,
	pages repository.KnowledgeBaseRepository,
	users repository.UserRepository,
	txManager repository.TxManager,
) {
	checkSpace := kb.NewCheckSpacePermissionUseCase(permissions)

	h := NewTicketHandler(
		checkSpace,
		ticket.NewCheckTicketPermissionUseCase(tickets, permissions),
		ticket.NewResolveTicketKeyUseCase(tickets),
		ticket.NewEnableTicketsForSpaceUseCase(tickets, txManager),
		ticket.NewCreateTicketUseCase(tickets),
		ticket.NewGetTicketUseCase(tickets),
		ticket.NewListTicketsUseCase(tickets),
		ticket.NewUpdateTicketUseCase(tickets),
		ticket.NewMoveTicketUseCase(tickets),
		ticket.NewArchiveTicketUseCase(tickets),
		ticket.NewRestoreTicketUseCase(tickets),
		ticket.NewChangeTicketStatusUseCase(tickets),
		ticket.NewChangeTicketParentUseCase(tickets),
		ticket.NewAssignTicketUseCase(tickets),
		ticket.NewUnassignTicketUseCase(tickets),
		ticket.NewListTicketHistoryUseCase(tickets),
	)
	sh := NewTicketStatusHandler(
		checkSpace,
		ticket.NewListTicketStatusesUseCase(tickets),
		ticket.NewCreateTicketStatusUseCase(tickets),
		ticket.NewUpdateTicketStatusUseCase(tickets),
		ticket.NewSetInitialTicketStatusUseCase(tickets),
		ticket.NewArchiveTicketStatusUseCase(tickets),
		ticket.NewRestoreTicketStatusUseCase(tickets),
	)
	th := NewTicketTypeHandler(
		checkSpace,
		ticket.NewListTicketTypesUseCase(tickets),
		ticket.NewCreateTicketTypeUseCase(tickets),
		ticket.NewUpdateTicketTypeUseCase(tickets),
		ticket.NewSetDefaultTicketTypeUseCase(tickets),
		ticket.NewArchiveTicketTypeUseCase(tickets),
		ticket.NewRestoreTicketTypeUseCase(tickets),
	)

	tkGroup := g.Group("", middleware.KnowledgeBaseWorkspace(
		kb.NewResolveWorkspaceUseCase(pages, permissions, users),
	))

	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/tickets/enable", h.Enable)
	tkGroup.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/tickets", h.List)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/tickets", h.Create)
	// 表示キー（例 FRESTYLE-12）からの解決。/tickets/:ticketId と衝突しないよう
	// /tickets/key/:key に独立させる（ticketId は UUID、key はハイフン入りの自由文字列）。
	tkGroup.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/tickets/key/:key", h.ResolveByKey)
	tkGroup.GET("/kb/workspaces/:workspaceSlug/tickets/:ticketId", h.Get)
	tkGroup.PUT("/kb/workspaces/:workspaceSlug/tickets/:ticketId", h.Update)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/tickets/:ticketId/move", h.Move)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/tickets/:ticketId/archive", h.Archive)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/tickets/:ticketId/restore", h.Restore)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/tickets/:ticketId/status", h.ChangeStatus)
	tkGroup.PUT("/kb/workspaces/:workspaceSlug/tickets/:ticketId/parent", h.ChangeParent)
	tkGroup.PUT("/kb/workspaces/:workspaceSlug/tickets/:ticketId/assignee", h.Assign)
	tkGroup.DELETE("/kb/workspaces/:workspaceSlug/tickets/:ticketId/assignee", h.Unassign)
	tkGroup.GET("/kb/workspaces/:workspaceSlug/tickets/:ticketId/history", h.History)

	// 状態マスタ（管理画面）。
	tkGroup.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses", sh.List)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses", sh.Create)
	tkGroup.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses/:statusId", sh.Update)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses/:statusId/set-initial", sh.SetInitial)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses/:statusId/archive", sh.Archive)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-statuses/:statusId/restore", sh.Restore)

	// 種別マスタ（管理画面）。
	tkGroup.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types", th.List)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types", th.Create)
	tkGroup.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types/:typeId", th.Update)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types/:typeId/set-default", th.SetDefault)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types/:typeId/archive", th.Archive)
	tkGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/ticket-types/:typeId/restore", th.Restore)
}
