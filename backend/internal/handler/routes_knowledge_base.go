package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerKnowledgeBaseRoutes はナレッジ基盤のページ操作と権限操作のエンドポイントを登録する。
//
// ワークスペースは URL の slug から middleware が解決するので、ルートはすべて
// /kb/workspaces/:workspaceSlug 以下に置き、その middleware を通す group に登録する
// （通し忘れたルートはテナント未確定のまま handler に入るため、group をここ 1 箇所に閉じる）。
func registerKnowledgeBaseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	registerKnowledgeBaseRoutesWith(
		g,
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewKnowledgeBasePermissionRepository(deps.db),
		persistence.NewWorkspaceProvisioner(deps.db),
	)
}

// registerKnowledgeBasePublicRoutes は認証不要のナレッジ基盤エンドポイントを登録する。
//
// ここに置いてよいのは「ログインしていない相手が使う」ものだけ。今のところ共有リンクの
// 検証 1 本で、認可はトークン（と任意のパスワード）そのものが担う。
func registerKnowledgeBasePublicRoutes(g *gin.RouterGroup, deps *routeDeps) {
	registerKnowledgeBasePublicRoutesWith(
		g,
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewKnowledgeBasePermissionRepository(deps.db),
	)
}

// registerKnowledgeBaseRoutesWith は repository を受け取ってルートと middleware を組み立てる。
// 本番の wiring とテストが同じ 1 箇所を通るようにするために切り出してある
// （テストがルート表を書き写すと、本番だけ middleware が抜けた配線ミスを見逃す）。
func registerKnowledgeBaseRoutesWith(
	g *gin.RouterGroup,
	pages repository.KnowledgeBaseRepository,
	permissions repository.KnowledgeBasePermissionRepository,
	provisioner repository.WorkspaceProvisioner,
) {
	h := NewKnowledgeBasePageHandler(
		usecase.NewCheckPagePermissionUseCase(permissions),
		usecase.NewCheckSpacePermissionUseCase(permissions),
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

	wh := NewKnowledgeBaseWorkspaceHandler(
		usecase.NewListMemberWorkspacesUseCase(permissions),
		usecase.NewCreateWorkspaceUseCase(provisioner),
		usecase.NewCheckWorkspacePermissionUseCase(permissions),
		usecase.NewCreateSpaceUseCase(pages),
	)

	// 権限操作 API の認可判定はこの 1 つの gate を共有する。
	// 「なぜ handler で認可を判定するのか」「なぜ super_admin を特別扱いしないのか」
	// 「なぜ拒否を 404 で揃えるのか」は kb_permission_gate.go の冒頭に書いてある。
	gate := newKbPermissionGate(
		usecase.NewCheckWorkspacePermissionUseCase(permissions),
		usecase.NewCheckSpacePermissionUseCase(permissions),
		usecase.NewFindPageUseCase(pages),
	)
	canRemoveAdmin := usecase.NewCanRemoveWorkspaceAdminUseCase(permissions)

	gh := NewKnowledgeBaseGrantHandler(
		gate,
		usecase.NewGrantWorkspaceRoleUseCase(permissions),
		usecase.NewRevokeWorkspaceRoleUseCase(permissions),
		usecase.NewGrantSpaceRoleUseCase(permissions),
		usecase.NewRevokeSpaceRoleUseCase(permissions),
		usecase.NewSetPageRestrictionUseCase(permissions),
		usecase.NewClearPageRestrictionUseCase(permissions),
		canRemoveAdmin,
	)

	mh := NewKnowledgeBaseMemberHandler(
		gate,
		usecase.NewAddWorkspaceMemberUseCase(permissions),
		usecase.NewRemoveWorkspaceMemberUseCase(permissions),
		usecase.NewCreatePrincipalGroupUseCase(permissions),
		usecase.NewAddGroupMemberUseCase(permissions),
		usecase.NewRemoveGroupMemberUseCase(permissions),
		usecase.NewEnsureSpaceEveryonePrincipalUseCase(permissions),
		canRemoveAdmin,
	)

	sh := NewKnowledgeBaseShareLinkHandler(
		gate,
		usecase.NewIssueShareLinkUseCase(permissions),
		usecase.NewRevokeShareLinkUseCase(permissions),
		usecase.NewListPageShareLinksUseCase(permissions),
		usecase.NewVerifyShareLinkUseCase(permissions),
	)

	// 所属ワークスペースの一覧と作成だけは middleware.KnowledgeBaseWorkspace を通さない。
	// あれは URL の slug から所属済みのワークスペースを確定させる middleware で、
	// 「どの slug を開けるのか」を知る前・そもそもワークスペースを作る前には使えない。
	// 認証（CurrentUser）は呼び出し元の group が既に通している。
	g.GET("/kb/workspaces", wh.List)
	// 作成は認証済みなら誰でも叩けて、slug はテナントをまたいで一意。
	// 上限が無いと 1 人で短い slug を取り尽くせてしまい、取り返す手段が運用の手作業しか無い。
	// 保有数の上限までは塞げないが、掴み取りの速度は他の作成系と同じ土俵に落とす。
	g.POST("/kb/workspaces", middleware.RateLimitPerMinute(10, 5), wh.Create)

	kb := g.Group("", middleware.KnowledgeBaseWorkspace(
		usecase.NewResolveWorkspaceUseCase(pages, permissions),
	))
	kb.POST("/kb/workspaces/:workspaceSlug/spaces", wh.CreateSpace)
	kb.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Tree)
	kb.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Create)
	kb.GET("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Get)
	kb.PATCH("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Rename)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/move", h.Move)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/archive", h.Archive)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/unarchive", h.Unarchive)
	kb.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/content", h.ReplaceContent)

	// ここから下が「権限そのものを変える」経路。すべて admin だけが通り、
	// 通らなかった要求は理由も対象の種類も伏せて 404 を返す（kb_permission_gate.go）。
	//
	// 既定の権限（grant）— ワークスペース全体とスペース単位の 2 段。
	kb.PUT("/kb/workspaces/:workspaceSlug/grants/:principalId", gh.GrantWorkspaceRole)
	kb.DELETE("/kb/workspaces/:workspaceSlug/grants/:principalId", gh.RevokeWorkspaceRole)
	kb.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId", gh.GrantSpaceRole)
	kb.DELETE("/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId", gh.RevokeSpaceRole)

	// ページ以下だけ既定を上書きする例外（restriction）。
	// URL が (ページ, 主体, ケイパビリティ) を指すのは、それが DB の主キーそのもので、
	// PUT / DELETE が同じ 1 行を指すため。
	kb.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/restrictions/:principalId/:capability", gh.SetPageRestriction)
	kb.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/restrictions/:principalId/:capability", gh.ClearPageRestriction)

	// 権限を張る相手（principals）の出し入れ。
	kb.PUT("/kb/workspaces/:workspaceSlug/members/:userId", mh.AddMember)
	kb.DELETE("/kb/workspaces/:workspaceSlug/members/:userId", mh.RemoveMember)
	kb.POST("/kb/workspaces/:workspaceSlug/groups", mh.CreateGroup)
	kb.PUT("/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId", mh.AddGroupMember)
	kb.DELETE("/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId", mh.RemoveGroupMember)
	kb.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/principals/everyone", mh.EnsureSpaceEveryone)

	// 共有リンク（発行・一覧・失効）。検証だけは未認証なので
	// registerKnowledgeBasePublicRoutesWith 側に置く。
	kb.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links", sh.ListShareLinks)
	kb.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links", sh.IssueShareLink)
	kb.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links/:shareLinkId", sh.RevokeShareLink)
}

// registerKnowledgeBasePublicRoutesWith は認証不要のルートを組み立てる。
//
// 共有リンクの検証は、認可をトークンそのものが担う唯一の経路。ログインしていない相手が
// 使うので middleware.KnowledgeBaseWorkspace（slug と所属からテナントを確定させる）を
// 通せず、ワークスペースはトークンから引いたリンクの側が持っている。
//
// トークンは 256 bit の乱数だが、パスワード付きリンクのパスワードは人が選ぶ短い値なので、
// 招待 token の検証（/invitations/accept/:token）と同じく IP 単位のレート制限をかける。
func registerKnowledgeBasePublicRoutesWith(
	g *gin.RouterGroup,
	pages repository.KnowledgeBaseRepository,
	permissions repository.KnowledgeBasePermissionRepository,
) {
	sh := NewKnowledgeBaseShareLinkHandler(
		newKbPermissionGate(
			usecase.NewCheckWorkspacePermissionUseCase(permissions),
			usecase.NewCheckSpacePermissionUseCase(permissions),
			usecase.NewFindPageUseCase(pages),
		),
		usecase.NewIssueShareLinkUseCase(permissions),
		usecase.NewRevokeShareLinkUseCase(permissions),
		usecase.NewListPageShareLinksUseCase(permissions),
		usecase.NewVerifyShareLinkUseCase(permissions),
	)
	g.POST("/kb/share-links/verify", middleware.RateLimitPerMinute(20, 10), sh.VerifyShareLink)
}
