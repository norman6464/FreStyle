package handler

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/infra/ratelimit"
	infraS3 "github.com/norman6464/FreStyle/backend/internal/infra/s3"
	"github.com/norman6464/FreStyle/backend/internal/usecase/comment"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
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
	// スペース作成はプライベートの導入でメンバー全員に開いた書き込みの口。
	// 人が手で作る回数としては十分に余裕があり、連打での作り散らかしだけを抑える。
	kbCreateSpacePerMinute = 20
	kbCreateSpaceBurst     = 10
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
		persistence.NewTxManager(deps.db),
		newKbImagePresignerOrFallback(deps),
	)
}

// newKbImagePresignerOrFallback は IMAGES_BUCKET 未設定なら stub にフォールバックする
// （bucket が最初から無い = 明示的にローカル開発用と分かる状態なので安全。rich-text 画像・
// profile 画像と同じバケットを kb/ prefix で共有する）。
//
// bucket が設定されているのに infraS3.NewPresigner が失敗する場合は fallback しない
// （CodeRabbit 指摘・段1b）。この場合は「本物の S3 を使うつもりだった」ことが bucket 名の
// 存在から明らかなので、黙って stub（未署名 URL）へ倒すと IssueImageUploadURL が 200 を
// 返し続け、クライアントは成功と誤認したまま S3 PUT だけが失敗する。config.Load の OIDC
// 必須化（「起動時に止める。通す側に倒すと誰も気づかない」）と同じ考え方で、ここも
// 起動を失敗させる。
func newKbImagePresignerOrFallback(deps *routeDeps) repository.KbImagePresigner {
	bucket := deps.cfg.S3.ImagesBucket
	if bucket == "" {
		log.Printf("[kb-image] IMAGES_BUCKET unset — using stub presigner (DEV)")
		return persistence.NewStubKbImagePresigner("stub-bucket")
	}
	pre, err := infraS3.NewPresigner(context.Background(), deps.cfg.S3.Region, bucket)
	if err != nil {
		log.Fatalf("[kb-image] IMAGES_BUCKET=%q is set but S3 presigner init failed: %v", bucket, err)
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
	txManager repository.TxManager,
	kbImagePresigner repository.KbImagePresigner,
) {
	// ReplacePageBlocksUseCase は本文保存の成功直後に versionRepo.CreateVersionIfDue を同じ
	// トランザクションで呼ぶ（FRESTYLE-433 段 3）ので、PageVersionHandler と同じ 1 つの
	// インスタンスを共有する（RestorePageVersionUseCase もこれをそのまま呼ぶ）。
	replaceBlocks := kb.NewReplacePageBlocksUseCase(pages, txManager, versions)
	h := NewKnowledgeBasePageHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
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
		kb.NewLookupUserNameUseCase(users),
		kb.NewIssuePageImageUploadURLUseCase(pages, kbImagePresigner),
		kb.NewIssuePageImageDownloadURLUseCase(pages, kbImagePresigner),
		kb.NewSetPageCoverUseCase(pages),
		kb.NewResolveCoverURLUseCase(kbImagePresigner),
	)

	// ページ全体へのコメント（FRESTYLE-432 段 2）。認可は CommentHandler 内で
	// CheckPagePermissionUseCase を直接使う（CanComment / CanView の判定は
	// requireCommentPermission / requirePagePermissionWith を参照）。
	ch := NewCommentHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		comment.NewCreateCommentThreadUseCase(comments, txManager),
		comment.NewAddCommentUseCase(comments),
		comment.NewListCommentThreadsUseCase(comments),
		comment.NewResolveCommentThreadUseCase(comments),
		comment.NewReopenCommentThreadUseCase(comments),
		kb.NewLookupUserNameUseCase(users),
	)

	// ページ本文の版（FRESTYLE-433 段 3）。一覧・単体取得は CapabilityView、
	// 作成（「版を残す」）・復元は CapabilityEdit（PageVersionHandler 内の各ハンドラ参照）。
	vh := NewPageVersionHandler(
		kb.NewCheckPagePermissionUseCase(permissions),
		kb.NewCreateExplicitPageVersionUseCase(versions, pages, txManager),
		kb.NewListPageVersionsUseCase(versions),
		kb.NewGetPageVersionUseCase(versions),
		kb.NewRestorePageVersionUseCase(versions, replaceBlocks),
		kb.NewLookupUserNameUseCase(users),
	)

	wh := NewKnowledgeBaseWorkspaceHandler(
		kb.NewListMemberWorkspacesUseCase(permissions),
		kb.NewJoinCompanyWorkspaceUseCase(permissions, users),
		kb.NewCreateWorkspaceUseCase(provisioner),
		kb.NewDeleteWorkspaceUseCase(pages),
		kb.NewCheckWorkspacePermissionUseCase(permissions),
		kb.NewCreateSpaceUseCase(pages, provisioner),
		kb.NewListViewableSpacesUseCase(permissions),
		kb.NewCheckSpacePermissionUseCase(permissions),
		kb.NewRenameSpaceUseCase(pages),
		kb.NewSearchViewablePagesUseCase(permissions),
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
		kb.NewAddWorkspaceMemberUseCase(permissions),
		kb.NewRemoveWorkspaceMemberUseCase(permissions),
		kb.NewCreatePrincipalGroupUseCase(permissions),
		kb.NewAddGroupMemberUseCase(permissions),
		kb.NewRemoveGroupMemberUseCase(permissions),
		kb.NewEnsureSpaceEveryonePrincipalUseCase(permissions),
		canRemoveAdmin,
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
	g.POST("/kb/workspaces", middleware.RateLimitPerMinute(10, 5), wh.Create)

	kbGroup := g.Group("", middleware.KnowledgeBaseWorkspace(
		kb.NewResolveWorkspaceUseCase(pages, permissions, users),
	))
	// スペースの一覧はワークスペースのメンバーなら誰でも叩ける（返る中身が権限で変わる）。
	// 作成と違って admin の gate を掛けないのは、これがサイドバーの入口だから。
	// 見せてよいスペースの選別は handler ではなく usecase 側のふるいが行う。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/spaces", wh.ListSpaces)
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
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/content", h.ReplaceContent)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/icon", h.SetIcon)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/icon", h.ClearIcon)
	// ページに閉じた画像の読み取り経路（FRESTYLE-368 段 1b）。
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/images/upload-url", h.IssueImageUploadURL)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/images/download-url", h.IssueImageDownloadURL)
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/pages/:pageId/cover", h.SetCover)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/pages/:pageId/cover", h.ClearCover)

	// ページ全体へのコメント（FRESTYLE-432 段 2）。一覧は CanView だけで許可し、
	// 作成・返信・解決・再開は CanComment を要求する（CommentHandler.requireCommentPermission）。
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads", ch.CreateThread)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads", ch.ListThreads)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/comments", ch.AddComment)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/resolve", ch.Resolve)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/comment-threads/:threadId/reopen", ch.Reopen)

	// ページ本文の版（FRESTYLE-433 段 3）。一覧・単体取得は CapabilityView（閲覧できれば
	// 誰でも読める）、「版を残す」・復元は CapabilityEdit を要求する（PageVersionHandler 参照）。
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/versions", vh.List)
	kbGroup.GET("/kb/workspaces/:workspaceSlug/pages/:pageId/versions/:seq", vh.Get)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/versions", vh.Create)
	kbGroup.POST("/kb/workspaces/:workspaceSlug/pages/:pageId/versions/:seq/restore", vh.Restore)

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
	// メンバー追加だけは回数に上限を置く。この口は users.id をそのまま受け取り、
	// 成功（200）と 404 の差でユーザーの実在が分かる。ワークスペースは認証済みなら誰でも
	// 作れて、作った本人が admin になるので、**全ログインユーザーが使える走査器**になっている。
	// 鍵はログイン中のユーザー（検証済み JWT 由来なので付け替えられない。IP は XFF で
	// 付け替えられるため鍵に使わない）。
	//
	// 走査そのものを塞ぐには「誰を招けるか」を会社などで絞る必要があり、それは
	// 権限モデルの外側の設計判断になる（同意なく他人を自分のワークスペースへ入れられる、
	// という別の問題も同じところに根がある）。ここで掛けるのは速度の頭打ちまで。
	kbGroup.PUT("/kb/workspaces/:workspaceSlug/members/:userId",
		middleware.RateLimitPerMinutePerUser(kbAddMemberPerMinute, kbAddMemberBurst), mh.AddMember)
	kbGroup.DELETE("/kb/workspaces/:workspaceSlug/members/:userId", mh.RemoveMember)
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
