package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/adapter/persistence"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	infraGCS "github.com/norman6464/frestyle/backend/internal/infra/gcs"
	"github.com/norman6464/frestyle/backend/internal/infra/ratelimit"
	"github.com/norman6464/frestyle/backend/internal/usecase/comment"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// 共有リンクの検証と、メンバー追加に掛ける上限の数値。
//
// 共有リンク: リンク 1 本あたり 1 分 10 回（短期は 5 回まで）。パスワードを打ち間違える
// 人の邪魔にはならず、総当たりの速度は 1 分 10 通りまで落ちる。鍵がリンクなので、
// この上限は要求元をいくら分散させても効く。同じリンクを持っている人どうしは
// 上限を共有するが、そもそもリンクを渡された者どうしなので実害は無い。
//
// メンバー追加: ユーザー 1 人あたり 1 分 30 回（短期は 10 回まで）。実運用の一括追加が
// 詰まらない程度に取りつつ、ユーザー ID 空間の走査は 1 分 30 件まで落ちる。
const (
	kbShareLinkVerifyPerMinute = 10
	kbShareLinkVerifyBurst     = 5
	kbAddMemberPerMinute       = 30
	kbAddMemberBurst           = 10
	// ワークスペース作成: ユーザー 1 人あたり 1 分 10 回（短期は 5 回まで）。IP ではなく
	// ユーザーを鍵にする — 未認証では叩けない経路なので JWT 由来のユーザー ID が
	// 必ず決まっており、これは詐称できない（IP は RealClientIP で詐称は防げても、
	// 同じ NAT の裏にいる無関係な相手を巻き添えにし得る。ここは 1 人の利用者が
	// 自動化で slug を先取りし続ける状況だけを抑えたいので、鍵は本人に絞れる）。
	kbCreateWorkspacePerMinute = 10
	kbCreateWorkspaceBurst     = 5
	// スペース作成はプライベートの導入でメンバー全員に開いた書き込みの口。
	// 人が手で作る回数としては十分に余裕があり、連打での作り散らかしだけを抑える。
	kbCreateSpacePerMinute = 20
	kbCreateSpaceBurst     = 10
	// 本文の JSON を解釈する 3 つの口（本文の保存・提案の作成・雛形からの作成）。
	// 1 本の要求で確保する記憶域は入力の上限と入れ子の段数で頭打ちにしてあるが、
	// 速さの側にも壁を置く。本文の保存は打っている最中に 0.8 秒ごとの自動保存が
	// 走るので、人が書き続けても詰まらない水準（1 分 120 回）に取る。
	// 提案と雛形からの作成は都度の操作なので、もっと絞ってよい。
	kbReplaceContentPerMinute = 120
	kbReplaceContentBurst     = 30
	kbParseDocPerMinute       = 30
	kbParseDocBurst           = 10
)

// registerKnowledgeBaseRoutes はナレッジのページ操作と権限操作のエンドポイントを登録する。
//
// ワークスペースは URL の slug から middleware が解決するので、ルートはすべて
// /kb/workspaces/:workspaceSlug 以下に置き、その middleware を通す group に登録する
// （通し忘れたルートはテナント未確定のまま handler に入るため、group をここ 1 箇所に閉じる）。
func registerKnowledgeBaseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	registerKnowledgeBaseRoutesWith(
		g,
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewKnowledgeBasePermissionRepository(deps.db),
		persistence.NewShareLinkRepository(deps.db),
		persistence.NewWorkspaceProvisioner(deps.db),
		persistence.NewUserRepository(deps.db),
		persistence.NewCommentRepository(deps.db),
		persistence.NewPageVersionRepository(deps.db),
		persistence.NewPageTemplateRepository(deps.db),
		persistence.NewPageSuggestionRepository(deps.db),
		persistence.NewTicketRepository(deps.db),
		persistence.NewTxManager(deps.db),
		newKbImagePresignerOrFallback(deps),
	)
}

// newKbImagePresignerOrFallback は IMAGES_BUCKET 未設定なら stub にフォールバックする
// （bucket が最初から無い = 明示的にローカル開発用と分かる状態なので安全。rich-text 画像・
// profile 画像と同じバケットを kb/ prefix で共有する）。
//
// bucket が設定されているのに infraGCS.NewPresigner が失敗する場合は fallback しない。
// この場合は「本物の Cloud Storage を使うつもりだった」ことが bucket 名の
// 存在から明らかなので、黙って stub（未署名 URL）へ倒すと IssueImageUploadURL が 200 を
// 返し続け、クライアントは成功と誤認したままアップロード PUT だけが失敗する。config.Load の
// OIDC 必須化（「起動時に止める。通す側に倒すと誰も気づかない」）と同じ考え方で、ここも
// 起動を失敗させる。
func newKbImagePresignerOrFallback(deps *routeDeps) repository.KbImagePresigner {
	bucket := deps.cfg.Images.Bucket
	if bucket == "" {
		log.Printf("[kb-image] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubKbImagePresigner("stub-bucket")
	}
	pre, err := infraGCS.NewPresigner(context.Background(), bucket)
	if err != nil {
		log.Fatalf("[kb-image] IMAGES_BUCKET=%q is set but GCS presigner init failed: %v", bucket, err)
	}
	return persistence.NewKbImagePresigner(pre)
}

// registerKnowledgeBasePublicRoutes は認証不要のナレッジエンドポイントを登録する。
//
// ここに置いてよいのは「ログインしていない相手が使う」ものだけ。今のところ共有リンクの
// 検証 1 本で、認可はトークン（と任意のパスワード）そのものが担う。
func registerKnowledgeBasePublicRoutes(g *gin.RouterGroup, deps *routeDeps) {
	registerKnowledgeBasePublicRoutesWith(
		g,
		persistence.NewKnowledgeBaseRepository(deps.db),
		persistence.NewKnowledgeBasePermissionRepository(deps.db),
		persistence.NewShareLinkRepository(deps.db),
	)
}

// registerKnowledgeBaseRoutesWith は repository を受け取ってルートと middleware を組み立てる。
// 本番の wiring とテストが同じ 1 箇所を通るようにするために切り出してある
// （テストがルート表を書き写すと、本番だけ middleware が抜けた配線ミスを見逃す）。
func registerKnowledgeBaseRoutesWith(
	g *gin.RouterGroup,
	pages repository.KnowledgeBaseRepository,
	permissions repository.KnowledgeBasePermissionRepository,
	shareLinks repository.ShareLinkRepository,
	provisioner repository.WorkspaceProvisioner,
	users repository.UserRepository,
	comments repository.CommentRepository,
	versions repository.PageVersionRepository,
	templates repository.PageTemplateRepository,
	suggestions repository.PageSuggestionRepository,
	tickets repository.TicketRepository,
	txManager repository.TxManager,
	kbImagePresigner repository.KbImagePresigner,
) {
	// ReplacePageBlocksUseCase は本文保存の成功直後に versionRepo.CreateVersionIfDue を同じ
	// トランザクションで呼ぶので、PageVersionHandler と同じ 1 つの
	// インスタンスを共有する（RestorePageVersionUseCase もこれをそのまま呼ぶ）。
	replaceBlocks := kb.NewReplacePageBlocksUseCase(pages, txManager, versions)
	h := NewKnowledgeBasePageHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		kb.NewCheckWorkspacePermissionUseCase(permissions),
		kb.NewResolvePageLocationUseCase(pages),
		kb.NewCheckSpacePermissionUseCase(permissions),
		kb.NewCanEditPageSubtreeUseCase(permissions),
		kb.NewListViewablePagesUseCase(permissions),
		kb.NewGetPageUseCase(pages),
		kb.NewFindPageUseCase(pages),
		kb.NewCreatePageUseCase(pages),
		kb.NewRenamePageUseCase(pages),
		kb.NewMovePageUseCase(pages),
		kb.NewArchivePageUseCase(pages),
		kb.NewUnarchivePageUseCase(pages),
		replaceBlocks,
		kb.NewResolvePageRefTitlesUseCase(permissions),
		kb.NewListViewableAncestorsUseCase(pages, permissions),
		kb.NewDeletePageUseCase(pages),
		kb.NewSetPageIconUseCase(pages),
		user.NewLookupUserDisplayUseCase(users),
		kb.NewIssuePageImageUploadURLUseCase(pages, kbImagePresigner),
		kb.NewIssuePageImageDownloadURLUseCase(pages, kbImagePresigner),
		kb.NewSetPageCoverUseCase(pages),
		kb.NewResolveCoverURLUseCase(kbImagePresigner),
		kb.NewListPageBacklinksUseCase(permissions),
		ticket.NewListTicketsReferencingPageUseCase(tickets),
	)

	// ページ全体へのコメント。認可は CommentHandler 内で
	// CheckPagePermissionUseCase を直接使う（CanComment / CanView の判定は
	// requireCommentPermission / requirePagePermissionWith を参照）。
	ch := NewCommentHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		comment.NewCreateCommentThreadUseCase(comments, txManager),
		comment.NewAddCommentUseCase(comments),
		comment.NewListCommentThreadsUseCase(comments),
		comment.NewResolveCommentThreadUseCase(comments),
		comment.NewReopenCommentThreadUseCase(comments),
		user.NewLookupUserDisplayUseCase(users),
	)

	// ページ本文の版。一覧・単体取得は CapabilityView、
	// 作成（「版を残す」）・復元は CapabilityEdit（PageVersionHandler 内の各ハンドラ参照）。
	vh := NewPageVersionHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		kb.NewCreateExplicitPageVersionUseCase(versions, pages, txManager),
		kb.NewListPageVersionsUseCase(versions),
		kb.NewGetPageVersionUseCase(versions),
		kb.NewRestorePageVersionUseCase(versions, replaceBlocks),
		user.NewLookupUserDisplayUseCase(users),
	)

	// ページの雛形。作成・削除はワークスペース全体への CanEdit、
	// 一覧はワークスペース所属者なら誰でも、使用（雛形からページを作る）は既存のページ作成
	// （h.Create）と全く同じ認可分岐で判定する（PageTemplateHandler 参照）。
	tplCheckSpace := kb.NewCheckSpacePermissionUseCase(permissions)
	th := NewPageTemplateHandler(
		kb.NewIsWorkspaceMemberUseCase(permissions),
		kb.NewCheckWorkspacePermissionUseCase(permissions),
		kb.NewCheckPagePermissionUseCase(permissions),
		tplCheckSpace,
		kb.NewListPageTemplatesUseCase(templates, tplCheckSpace),
		kb.NewCreateTemplateFromPageUseCase(pages, templates, tplCheckSpace),
		kb.NewDeletePageTemplateUseCase(templates, tplCheckSpace),
		kb.NewCreatePageFromTemplateUseCase(templates, tplCheckSpace, kb.NewCreatePageUseCase(pages), replaceBlocks, kb.NewDeletePageUseCase(pages)),
	)

	// 提案。作成は CanComment、一覧の閲覧は CanView、採用・却下は CanEdit
	// （PageSuggestionHandler 参照）。採用は本文保存の成功直後に版を切る通常の保存経路と
	// 同じ replaceBlocks インスタンスを使い回す（インスタンスを複数持つと版のトランザクション境界が
	// 揃わなくなるため）。
	sgh := NewPageSuggestionHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		kb.NewCreateSuggestionUseCase(pages, versions, suggestions),
		kb.NewListOpenPageSuggestionsUseCase(suggestions),
		kb.NewAcceptPageSuggestionUseCase(suggestions, versions, replaceBlocks, txManager),
		kb.NewRejectPageSuggestionUseCase(suggestions),
		kb.NewGetPageVersionUseCase(versions),
		user.NewLookupUserDisplayUseCase(users),
	)

	wh := NewKnowledgeBaseWorkspaceHandler(
		kb.NewListMemberWorkspacesUseCase(permissions),
		kb.NewCreateWorkspaceUseCase(provisioner),
		kb.NewDeleteWorkspaceUseCase(pages),
		kb.NewCheckWorkspacePermissionUseCase(permissions),
		kb.NewCreateSpaceUseCase(pages, provisioner),
		kb.NewListViewableSpacesUseCase(permissions),
		kb.NewCheckSpacePermissionUseCase(permissions),
		kb.NewRenameSpaceUseCase(pages),
		kb.NewSearchViewablePagesUseCase(permissions),
		kb.NewListWorkspaceMembersUseCase(permissions),
		kb.NewListWorkspaceMembersForAdminUseCase(permissions),
		kb.NewListMembershipEventsUseCase(permissions),
		user.NewLookupUserDisplayUseCase(users),
	)

	// 権限操作 API の認可判定はこの 1 つの gate を共有する。
	// 「なぜ handler で認可を判定するのか」「なぜ super_admin を特別扱いしないのか」
	// 「なぜ拒否を 404 で揃えるのか」は kb_permission_gate.go の冒頭に書いてある。
	gate := newKbPermissionGate(
		kb.NewCheckWorkspacePermissionUseCase(permissions),
		kb.NewCheckSpacePermissionUseCase(permissions),
		kb.NewCheckPagePermissionUseCase(permissions),
	)
	canRemoveAdmin := kb.NewCanRemoveWorkspaceAdminUseCase(permissions)

	gh := NewKnowledgeBaseGrantHandler(
		gate,
		kb.NewGrantWorkspaceRoleUseCase(permissions),
		kb.NewRevokeWorkspaceRoleUseCase(permissions),
		kb.NewGrantSpaceRoleUseCase(permissions),
		kb.NewRevokeSpaceRoleUseCase(permissions),
		kb.NewGrantPageRoleUseCase(permissions),
		kb.NewRevokePageRoleUseCase(permissions),
		kb.NewListPageGrantsUseCase(permissions),
		kb.NewListGrantablePrincipalsUseCase(permissions),
		canRemoveAdmin,
	)

	mh := NewKnowledgeBaseMemberHandler(
		gate,
		kb.NewInviteWorkspaceMemberUseCase(permissions),
		kb.NewRemoveWorkspaceMemberUseCase(permissions),
		kb.NewCreatePrincipalGroupUseCase(permissions),
		kb.NewAddGroupMemberUseCase(permissions),
		kb.NewRemoveGroupMemberUseCase(permissions),
		kb.NewEnsureSpaceEveryonePrincipalUseCase(permissions),
		canRemoveAdmin,
		user.NewSetUserActiveUseCase(users, permissions, txManager),
	)

	// この group には検証（Verify）を登録しないので、渡す limiter は使われない。
	// それでも組み立てるのは、handler の組み立て方をここと公開 group で揃えるため
	// （片方だけ nil を渡す形にすると、うっかり検証を認証済み側へ生やしたときに
	// 上限が無いまま動く）。
	sh := NewKnowledgeBaseShareLinkHandler(
		gate,
		kb.NewIssueShareLinkUseCase(shareLinks),
		kb.NewRevokeShareLinkUseCase(shareLinks),
		kb.NewListPageShareLinksUseCase(shareLinks),
		kb.NewVerifyShareLinkUseCase(shareLinks),
		ratelimit.New(kbShareLinkVerifyPerMinute, kbShareLinkVerifyBurst),
	)

	// 所属ワークスペースの一覧と作成だけは middleware.KnowledgeBaseWorkspace を通さない。
	// あれは URL の slug から所属済みのワークスペースを確定させる middleware で、
	// 「どの slug を開けるのか」を知る前・そもそもワークスペースを作る前には使えない。
	// 認証（CurrentUser）は呼び出し元の group が既に通している。
	g.GET("/kb/workspaces", wh.List)
	// /p/{pageId} の解決。URL にテナントを持たないため slug の middleware は通せない
	// （権限判定は handler の中で、解決した workspace に対して必ず行う）。
	g.GET("/kb/pages/:pageId", h.ResolveByID)
	// 作成は認証済みなら誰でも叩けて、slug はテナントをまたいで一意。
	// 上限が無いと 1 人で短い slug を取り尽くせてしまい、取り返す手段が運用の手作業しか無い。
	// 保有数の上限までは塞げないが、掴み取りの速度は他の作成系と同じ土俵に落とす。
	// 鍵はユーザー単位（kbCreateWorkspacePerMinute の doc 参照）。
	g.POST("/kb/workspaces", middleware.RateLimitPerMinutePerUser(kbCreateWorkspacePerMinute, kbCreateWorkspaceBurst), wh.Create)

	// 自分宛の招待（段 2）。受諾するまで所属していないので、ここも
	// middleware.KnowledgeBaseWorkspace を通さない（所属済みしか通さないため）。
	ih := NewKnowledgeBaseInvitationHandler(
		kb.NewListMyWorkspaceInvitationsUseCase(permissions),
		kb.NewAcceptWorkspaceInvitationUseCase(pages, permissions),
		kb.NewDeclineWorkspaceInvitationUseCase(pages, permissions),
	)
	g.GET("/kb/invitations", ih.List)
	g.POST("/kb/invitations/:workspaceSlug/accept", ih.Accept)
	g.POST("/kb/invitations/:workspaceSlug/decline", ih.Decline)

	kbGroup := g.Group("", middleware.KnowledgeBaseWorkspace(
		kb.NewResolveWorkspaceUseCase(pages, permissions),
	))
	// スペースの一覧はワークスペースのメンバーなら誰でも叩ける（返る中身が権限で変わる）。
	// 作成と違って admin の gate を掛けないのは、これがサイドバーの入口だから。
	// 見せてよいスペースの選別は handler ではなく usecase 側のふるいが行う。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/spaces", wh.ListSpaces)
	// ワークスペースの人の一覧。所属していれば誰でも叩ける（担当の表示名・発言での名指しに使う）。
	// 権限を張る相手を選ぶ /pages/:pageId/principals とは別の口（あちらはページの管理権限が要る）。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/members", wh.ListMembers)
	// 所属・権限の変更履歴（段 6・監査）。admin だけが見られる（handler 内で CanManage を確認）。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/membership-events", wh.ListMembershipEvents)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/admin/members", wh.ListMembersForAdmin)
	// ワークスペースの削除（配下ごと・戻せない）。会社のワークスペースは SQL 側で守る。
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug", wh.Delete)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/spaces",
		middleware.RateLimitPerMinutePerUser(kbCreateSpacePerMinute, kbCreateSpaceBurst), wh.CreateSpace)
	kbGroup.PATCH("/kb/workspaces/:workspaceSlug/spaces/:spaceId", wh.RenameSpace)
	// 検索は /pages/:pageId と衝突しないよう /search を独立させる。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/search", wh.SearchPages)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Tree)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages", h.Create)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Get)
	kbGroup.PATCH("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Rename)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId", h.Delete)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/move", h.Move)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/archive", h.Archive)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/unarchive", h.Unarchive)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/content",
		middleware.RateLimitPerMinutePerUser(kbReplaceContentPerMinute, kbReplaceContentBurst), h.ReplaceContent)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/icon", h.SetIcon)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/icon", h.ClearIcon)
	// ページに閉じた画像の読み取り経路。
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/images/upload-url", h.IssueImageUploadURL)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/images/download-url", h.IssueImageDownloadURL)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/cover", h.SetCover)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/cover", h.ClearCover)
	// 逆リンク: このページを参照しているページの一覧。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/backlinks", h.Backlinks)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/ticket-backlinks", h.TicketBacklinks)

	// ページ全体へのコメント。一覧は CanView だけで許可し、
	// 作成・返信・解決・再開は CanComment を要求する（CommentHandler.requireCommentPermission）。
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads", ch.CreateThread)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads", ch.ListThreads)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/comments", ch.AddComment)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/resolve", ch.Resolve)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/reopen", ch.Reopen)

	// ページ本文の版。一覧・単体取得は CapabilityView（閲覧できれば
	// 誰でも読める）、「版を残す」・復元は CapabilityEdit を要求する（PageVersionHandler 参照）。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/versions", vh.List)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/versions/:seq", vh.Get)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/versions", vh.Create)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/versions/:seq/restore", vh.Restore)

	// ページの雛形。一覧はワークスペース所属者なら誰でも、
	// 作成（そのページを雛形として保存）・削除はワークスペース全体への CanEdit を要求する
	// （PageTemplateHandler 参照）。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/templates", th.List)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/templates", th.CreateFromPage)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/templates/:templateId", th.Delete)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/spaces/:spaceId/pages/from-template",
		middleware.RateLimitPerMinutePerUser(kbParseDocPerMinute, kbParseDocBurst), th.CreatePage)

	// 提案。作成は CanComment、一覧の閲覧は CanView、採用・却下は CanEdit
	// を要求する（PageSuggestionHandler 参照）。
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/suggestions",
		middleware.RateLimitPerMinutePerUser(kbParseDocPerMinute, kbParseDocBurst), sgh.Create)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/suggestions", sgh.ListOpen)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/suggestions/:suggestionId/accept", sgh.Accept)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/suggestions/:suggestionId/reject", sgh.Reject)

	// ここから下が「権限そのものを変える」経路。すべて admin だけが通り、
	// 通らなかった要求は理由も対象の種類も伏せて 404 を返す（kb_permission_gate.go）。
	//
	// 既定の権限（grant）— ワークスペース全体とスペース単位の 2 段。
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/grants/:principalId", gh.GrantWorkspaceRole)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/grants/:principalId", gh.RevokeWorkspaceRole)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId", gh.GrantSpaceRole)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/spaces/:spaceId/grants/:principalId", gh.RevokeSpaceRole)
	// ページ単位の grant（既定の 3 段目）。このページとその子孫に効く。
	// 一覧が返すのはこの段で足した行だけで、上の段や祖先から届いている相手は含まない。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/grants", gh.ListPageGrants)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/grants/:principalId", gh.GrantPageRole)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/grants/:principalId", gh.RevokePageRole)
	// 権限を張れる相手（画面の相手選び）。認可はページ単位で、返る中身はワークスペース全体。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/principals", gh.ListGrantablePrincipals)

	// 権限を張る相手（principals）の出し入れ。
	// メンバー招待だけは回数に上限を置く。この口は users.id をそのまま受け取るため、
	// 招待の成否（204 と 404 の差）でユーザーの実在は分かる（列挙そのものは完全には
	// 塞げていない）。ただし段 2 より前と違い、招待しただけでは principal も権限も
	// 一切発生しない（本人が /kb/invitations/:workspaceSlug/accept を呼ぶまで所属しない。
	// FRESTYLE-486 の主眼だった「同意なく他人を追加できる」問題はこちらで塞いでいる）。
	// 鍵はログイン中のユーザー（検証済み JWT 由来なので付け替えられない。IP は XFF で
	// 付け替えられるため鍵に使わない）。
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/members/:userId",
		middleware.RateLimitPerMinutePerUser(kbAddMemberPerMinute, kbAddMemberBurst), mh.InviteMember)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/members/:userId", mh.RemoveMember)
	// アカウントの停止・復帰（段 7）。効果は全ワークスペースに及ぶが、実行できるのは
	// 対象が現に所属するこのワークスペースの admin だけ（kb_member_handler.go の
	// KnowledgeBaseMemberHandler.Suspend の doc 参照）。
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/members/:userId/suspend", mh.Suspend)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/members/:userId/restore", mh.Restore)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/groups", mh.CreateGroup)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId", mh.AddGroupMember)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/groups/:groupPrincipalId/members/:userId", mh.RemoveGroupMember)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/spaces/:spaceId/principals/everyone", mh.EnsureSpaceEveryone)

	// 共有リンク（発行・一覧・失効）。発行と失効は「誰が見られるか」を変える操作。
	// 検証だけは未認証なので registerKnowledgeBasePublicRoutesWith 側に置く。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links", sh.ListShareLinks)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links", sh.IssueShareLink)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/share-links/:shareLinkId", sh.RevokeShareLink)
}

// registerKnowledgeBasePublicRoutesWith は認証不要のルートを組み立てる。
//
// 共有リンクの検証は、認可をトークンそのものが担う唯一の経路。ログインしていない相手が
// 使うので middleware.KnowledgeBaseWorkspace（slug と所属からテナントを確定させる）を
// 通せず、ワークスペースはトークンから引いたリンクの側が持っている。
//
// トークンは 256 bit の乱数だが、パスワード付きリンクのパスワードは人が選ぶ短い値なので、
// 試行回数に上限をかける。鍵は IP ではなく**リンクそのもの**で、IP を変えても頭打ちになる
// （kbShareLinkAttemptKey の doc に理由がある）。IP 単位の上限も重ねるが、あれは
// 攻撃者が鍵を変えられるので、単独では総当たりの歯止めにならない。
func registerKnowledgeBasePublicRoutesWith(
	g *gin.RouterGroup,
	pages repository.KnowledgeBaseRepository,
	permissions repository.KnowledgeBasePermissionRepository,
	shareLinks repository.ShareLinkRepository,
) {
	sh := NewKnowledgeBaseShareLinkHandler(
		newKbPermissionGate(
			kb.NewCheckWorkspacePermissionUseCase(permissions),
			kb.NewCheckSpacePermissionUseCase(permissions),
			kb.NewCheckPagePermissionUseCase(permissions),
		),
		kb.NewIssueShareLinkUseCase(shareLinks),
		kb.NewRevokeShareLinkUseCase(shareLinks),
		kb.NewListPageShareLinksUseCase(shareLinks),
		kb.NewVerifyShareLinkUseCase(shareLinks),
		ratelimit.New(kbShareLinkVerifyPerMinute, kbShareLinkVerifyBurst),
	)
	// 上限は 2 段。**本命は handler 側のリンク 1 本あたりの上限**で、こちらの IP 単位は
	// 素直な大量アクセスを薄めるだけの層（XFF を詐称すれば鍵が変わるので、これだけでは
	// パスワードの総当たりを止められない）。詳細は kbShareLinkAttemptKey の doc。
	g.POST("/kb/share-links/verify", middleware.RateLimitPerMinute(20, 10), sh.VerifyShareLink)
}
