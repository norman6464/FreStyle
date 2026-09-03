package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// KnowledgeBasePageHandler はノートのページ操作を受ける。
//
// ワークスペースはリクエストからは受け取らず、middleware.KnowledgeBaseWorkspace が
// URL の slug と principals から確定させたものを context から取る。
type KnowledgeBasePageHandler struct {
	check          *usecase.CheckPagePermissionUseCase
	resolve        *usecase.ResolvePageLocationUseCase
	checkSpace     *usecase.CheckSpacePermissionUseCase
	canEditSubtree *usecase.CanEditPageSubtreeUseCase
	listViewable   *usecase.ListViewablePagesUseCase
	get            *usecase.GetPageUseCase
	findPage       *usecase.FindPageUseCase
	create         *usecase.CreatePageUseCase
	rename         *usecase.RenamePageUseCase
	move           *usecase.MovePageUseCase
	archive        *usecase.ArchivePageUseCase
	unarchive      *usecase.UnarchivePageUseCase
	replaceBlocks  *usecase.ReplacePageBlocksUseCase
	resolveRefs    *usecase.ResolvePageRefTitlesUseCase
	ancestors      *usecase.ListViewableAncestorsUseCase
	deletePage     *usecase.DeletePageUseCase
}

// NewKnowledgeBasePageHandler は KnowledgeBasePageHandler を組み立てる。
func NewKnowledgeBasePageHandler(
	check *usecase.CheckPagePermissionUseCase,
	resolve *usecase.ResolvePageLocationUseCase,
	checkSpace *usecase.CheckSpacePermissionUseCase,
	canEditSubtree *usecase.CanEditPageSubtreeUseCase,
	listViewable *usecase.ListViewablePagesUseCase,
	get *usecase.GetPageUseCase,
	findPage *usecase.FindPageUseCase,
	create *usecase.CreatePageUseCase,
	rename *usecase.RenamePageUseCase,
	move *usecase.MovePageUseCase,
	archive *usecase.ArchivePageUseCase,
	unarchive *usecase.UnarchivePageUseCase,
	replaceBlocks *usecase.ReplacePageBlocksUseCase,
	resolveRefs *usecase.ResolvePageRefTitlesUseCase,
	ancestors *usecase.ListViewableAncestorsUseCase,
	deletePage *usecase.DeletePageUseCase,
) *KnowledgeBasePageHandler {
	return &KnowledgeBasePageHandler{
		check:          check,
		resolve:        resolve,
		checkSpace:     checkSpace,
		canEditSubtree: canEditSubtree,
		listViewable:   listViewable,
		get:            get,
		findPage:       findPage,
		create:         create,
		rename:         rename,
		move:           move,
		archive:        archive,
		unarchive:      unarchive,
		replaceBlocks:  replaceBlocks,
		resolveRefs:    resolveRefs,
		ancestors:      ancestors,
		deletePage:     deletePage,
	}
}

// maxKnowledgeBaseBodyBytes はページ本文 API のボディ上限。bind 前に切って巨大ボディの
// 全読み込みを防ぐ（本文は ProseMirror の JSON なので文書 API と同じ桁で足りる）。
const maxKnowledgeBaseBodyBytes = (1 << 20) + (64 << 10) // 1 MiB + 64 KiB

// kbPageResponse はページ 1 件の返却形。
//
// workspaceId は載せない。ワークスペースは URL の slug と現在のユーザーの所属から決まる
// サーバ側の関心事で、クライアントが指定に使う値ではないため（内部 UUID を配って
// 「次はこれを送ればいい」と誤解させない）。
// position（並び順のキー）は**返さない**。
//
// 分数インデックスの整数部は末尾追加のたびに 1 ずつ増えるので、a0 と a3 が見えて a1 a2 が
// 見えなければ、その間に 2 枚あることがそのまま読める。hasHiddenChildren を有無に落として
// 枚数を伏せた意味が、この 1 項目で消える。
//
// 返さなくても困らない。並び順は backend が position 順に並べた配列として渡しており、
// 移動 API（kbMovePageRequest）は parentId だけを受けて位置はサーバが決める。
// クライアントがキーの値を必要とする経路が無い。
//
// 将来「A と B の間に入れる」が要るようになっても、渡すのは隣のページの **ID** にする。
// 生のキーを往復させると、この漏れが戻ってくる。
type kbPageResponse struct {
	ID              string     `json:"id"              example:"0198a000-0000-7000-8000-000000000003"`
	SpaceID         string     `json:"spaceId"         example:"0198a000-0000-7000-8000-000000000002"`
	ParentID        *string    `json:"parentId,omitempty"`
	Title           string     `json:"title"           example:"設計メモ"`
	CreatedByUserID uint64     `json:"createdByUserId" example:"42"`
	ArchivedAt      *time.Time `json:"archivedAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func toKbPageResponse(p *domain.Page) kbPageResponse {
	return kbPageResponse{
		ID:              p.ID,
		SpaceID:         p.SpaceID,
		ParentID:        p.ParentID,
		Title:           p.Title,
		CreatedByUserID: p.CreatedByUserID,
		ArchivedAt:      p.ArchivedAt,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
	}
}

// kbPageTreeResponse はツリーの 1 ノード（子を再帰的に含む）。
type kbPageTreeResponse struct {
	Page     kbPageResponse       `json:"page"`
	Children []kbPageTreeResponse `json:"children"`
	// HasHiddenChildren はこのページの直下に、閲覧できないページが在るか。
	// 枚数も題名も出さない。理由は ListViewablePagesOutput の doc に書いてある。
	HasHiddenChildren bool `json:"hasHiddenChildren" example:"false"`
	// ParentArchived は親がアーカイブ済みか。アーカイブ済みの一覧でだけ意味を持つ
	// （現役の一覧では常に false）。**事実であって判断ではない** — 復帰できるかの規則は
	// 「親がアーカイブ中なら断る」で、それを持つのは UnarchivePageUseCase。
	ParentArchived bool `json:"parentArchived" example:"false"`
}

// kbPageTreeRootResponse はツリー取得の応答全体。
//
// 配列ではなく object にしてあるのは、**スペース直下**にも同じ印を載せる場所が要るため。
// 「見えない子が居る」ことを段ごとに示す以上、いちばん上の段だけ示せないのは筋が通らない。
type kbPageTreeRootResponse struct {
	Pages []kbPageTreeResponse `json:"pages"`
	// HasHiddenChildren はスペース直下に、閲覧できないページが在るか。
	// 1 件も見えないスペースでは必ず false（存在しないスペースと撃ち分けないため）。
	HasHiddenChildren bool `json:"hasHiddenChildren" example:"false"`
}

func toKbPageTreeResponse(nodes []*usecase.PageTreeNode, hidden, parentArchived map[string]bool) []kbPageTreeResponse {
	out := make([]kbPageTreeResponse, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, kbPageTreeResponse{
			Page:              toKbPageResponse(&n.Page),
			Children:          toKbPageTreeResponse(n.Children, hidden, parentArchived),
			HasHiddenChildren: hidden[n.Page.ID],
			ParentArchived:    parentArchived[n.Page.ID],
		})
	}
	return out
}

// kbPageDocResponse はページのメタ情報と本文（ProseMirror doc）の組。
type kbPageDocResponse struct {
	Page kbPageResponse  `json:"page"`
	Doc  json.RawMessage `json:"doc"`
}

// respondKnowledgeBaseErr は usecase / repository のセンチネルを HTTP ステータスへ対応づける。
//
// 「存在しない」と「見る権限が無い」は必ず同じ 404 + 同じ本文にする。片方だけ別の応答を
// 返すと、ID を総当たりするだけで隠したページの実在が分かってしまう。
func respondKnowledgeBaseErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrPageNotFound),
		errors.Is(err, repository.ErrSpaceNotFound),
		errors.Is(err, repository.ErrWorkspaceNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, usecase.ErrPagePermissionDenied):
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
	case errors.Is(err, usecase.ErrPageArchived):
		c.JSON(http.StatusConflict, errorResponse{Error: "page_archived"})
	case errors.Is(err, usecase.ErrPageParentArchived):
		c.JSON(http.StatusConflict, errorResponse{Error: "parent_archived"})
	case errors.Is(err, usecase.ErrPageCycle):
		c.JSON(http.StatusConflict, errorResponse{Error: "page_cycle"})
	case errors.Is(err, repository.ErrPageMoveVoidsSpaceGrant):
		// 「今の権限設定のままでは移せない」という業務上の衝突であって、サーバの故障ではない。
		// 既にアーカイブ済み・循環と同じ 409 に揃える（どれも「リクエスト自体は正しいが、
		// 対象の現在の状態と両立しない」）。500 で返すと、クライアントは DB 障害と区別できず
		// 再試行してよいものと誤解する（何度試しても同じ結果になる）。
		c.JSON(http.StatusConflict, errorResponse{Error: "space_grant_voided"})
	case errors.Is(err, repository.ErrWorkspaceSlugTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "slug_taken"})
	case errors.Is(err, repository.ErrSpaceKeyTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "space_key_taken"})
	case errors.Is(err, repository.ErrWorkspaceHasMembers):
		// 人が居るワークスペースは消せない。実在は既に知っている相手なので理由を返す。
		c.JSON(http.StatusForbidden, errorResponse{Error: "workspace_has_members"})
	case errors.Is(err, repository.ErrPrincipalNotFound):
		// 所属が確かめられた後に外された場合にここへ来る。権限の拒否なので、
		// ほかの拒否と同じ 404 に畳む（500 にすると再試行してよいと誤解される）。
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, usecase.ErrInvalidWorkspaceSlug),
		errors.Is(err, usecase.ErrInvalidSpaceKey),
		errors.Is(err, usecase.ErrInvalidSpaceVisibility),
		errors.Is(err, usecase.ErrInvalidName):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
	case errors.Is(err, usecase.ErrPageParentSpaceMismatch):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "parent_space_mismatch"})
	case errors.Is(err, usecase.ErrPageAnchorNotSibling):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "anchor_not_sibling"})
	case errors.Is(err, usecase.ErrPageDocInvalid), errors.Is(err, usecase.ErrPageDocUnknownNodeType):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_document"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}

// kbRequestScope は 1 リクエストで確定したテナントと呼び出し元。
type kbRequestScope struct {
	workspaceID string
	userID      uint64
}

// scope は middleware が確定させたワークスペースと current user を取り出す。
// どちらか欠けていればレスポンスを書いて ok=false を返す（handler 側で ID を組み立てない）。
func kbScope(c *gin.Context) (kbRequestScope, bool) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return kbRequestScope{}, false
	}
	workspaceID := middleware.KnowledgeBaseWorkspaceIDOrEmpty(c)
	if workspaceID == "" {
		// middleware.KnowledgeBaseWorkspace を通さずに登録されたルート = 配線ミス。
		// テナント未確定のまま処理を続けると全テナントに触れてしまうので必ず落とす。
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return kbRequestScope{}, false
	}
	return kbRequestScope{workspaceID: workspaceID, userID: uid}, true
}

// requirePagePermission は 1 ページの実効権限を確かめる。満たさなければレスポンスを書いて false を返す。
// ページを名指しする経路はすべてこれを通す（判定規則は usecase / domain 側にあり、ここには写さない）。
func (h *KnowledgeBasePageHandler) requirePagePermission(
	c *gin.Context, scope kbRequestScope, pageID string, capability domain.Capability,
) bool {
	perm, err := h.check.Execute(c.Request.Context(), usecase.CheckPagePermissionInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return false
	}
	if !perm.CanView {
		// 閲覧できない相手にはページの実在を教えない（存在しない ID と同じ応答）。
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return false
	}
	if !perm.Allows(capability) {
		// ここに来る相手は閲覧できる = 実在を既に知っているので、403 で理由を返してよい。
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// requireSpacePermission はスペース 1 つの実効権限を確かめる。満たさなければレスポンスを
// 書いて false を返す。
//
// **ページを名指しする経路でこれを使ってはいけない。** スペースの判定はページ付与
// （page_grants）を見ておらず、祖先のページで足された役割を取りこぼす。
// 使ってよいのは対象がまだ存在しない操作（スペース直下へのページ作成）だけで、
// 親を持つ作成は requirePagePermission を通す。
func (h *KnowledgeBasePageHandler) requireSpacePermission(
	c *gin.Context, scope kbRequestScope, spaceID string, capability domain.Capability,
) bool {
	perm, err := h.checkSpace.Execute(c.Request.Context(), usecase.CheckSpacePermissionInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return false
	}
	if !perm.CanView {
		// 中身を 1 つも見られない相手にはスペースの実在を教えない
		// （ツリー取得が「無いスペース」と「空のスペース」を撃ち分けないのと同じ扱い）。
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return false
	}
	if !perm.Allows(capability) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// requireSubtreeEditPermission はページと全子孫の編集権限を確かめる。満たさなければ
// レスポンスを書いて false を返す。子孫ごと影響が及ぶ操作（アーカイブ / 復帰 / 移動）が通す。
//
// いまの権限モデルでは役割は木を下るほど弱くならないので、この検査が断ることは無い
// （理由は CanEditPageSubtreeUseCase の doc）。事実を集めるクエリの回帰を捕まえる
// 最後の網として残してある。部分的にアーカイブして逃げる手も採れない —
// アーカイブ済みの親の下に現役の子が残るとツリーに現れない迷子ページになり、
// 復帰の前提（親から順に戻す）も壊れる。
// 全部できるか、何もしないかの二択なので、フェイルクローズ側に倒して断る。
//
// 引き換えに、断ること自体が「この下に触れないページがある」という粒度の粗い信号になる
// （どのページかは分からない）。ページの実在を隠す規則との衝突は承知のうえで、
// 見えないページを黙って書き換えられる方を重く見た。
//
// 問い合わせはページ数によらず 1 回（CanEditPageSubtreeUseCase がサブツリーの事実を
// まとめて集め、domain.ResolvePagePermission に 1 ページずつ通す）。判定規則を
// ここへ写経しないこと — 写せば「直接触ると 403 なのに経由すると通る」が復活する。
func (h *KnowledgeBasePageHandler) requireSubtreeEditPermission(
	c *gin.Context, scope kbRequestScope, pageID string,
) bool {
	ok, err := h.canEditSubtree.Execute(c.Request.Context(), usecase.CanEditPageSubtreeInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return false
	}
	if !ok {
		c.JSON(http.StatusForbidden, errorResponse{Error: "subtree_forbidden"})
		return false
	}
	return true
}

// Tree はスペース配下の、そのユーザーが閲覧できるページを木構造で返す。
func (h *KnowledgeBasePageHandler) Tree(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	// archived=true でアーカイブ済みの一覧に切り替える。**別の口にしない**のは、
	// 権限の見方が現役とまったく同じだから（違うのは対象の絞り込みだけ）。
	// 口を分けると、片方だけ直して食い違う形をわざわざ作ることになる。
	archived := c.Query("archived") == "true"
	// ページごとに権限を引くと N+1 になるので、一覧はまとめて 1 回で解決する。
	viewable, err := h.listViewable.Execute(c.Request.Context(), usecase.ListViewablePagesInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     c.Param("spaceId"),
		UserID:      scope.userID,
		Archived:    archived,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// スペースの実在確認はしない。「無いスペース」と「中身が 1 件も見えないスペース」を
	// 撃ち分けると、スペース ID の総当たりで実在が分かってしまうため、どちらも空配列にする。
	//
	// 見えない親の子は根へ昇格させない（PageTreeOrphanHidden）。昇格させると、隠した親の
	// タイトルは伏せたまま「その下に何かがある」ことだけがツリーの形から漏れる。
	// 孤児（親が一覧に無いページ）の扱いは、どちらの一覧かで変える。
	//
	// 現役では落とす。昇格させると「見えない親の下に何かがある」ことがツリーの形から漏れる。
	//
	// アーカイブ済みでは根へ昇格させる。**アーカイブの根は必ず孤児になる**（その親は
	// 現役なので、この一覧には入らない）ため、落とすと 1 件も出なくなる。
	// 昇格させても現役のような漏れは起きない — 昇格した行が「親が現役の根」なのか
	// 「親もアーカイブ済みだが自分には見えない」のかは、応答から区別が付かない。
	// 前者だけが復帰できるので、その違いは parentArchived という事実として返す。
	policy := usecase.PageTreeOrphanHidden
	if archived {
		policy = usecase.PageTreeOrphanAsRoot
	}
	tree := usecase.BuildPageTree(viewable.Pages, policy)
	c.JSON(http.StatusOK, kbPageTreeRootResponse{
		Pages:             toKbPageTreeResponse(tree, viewable.HasHiddenChildren, viewable.ParentArchived),
		HasHiddenChildren: viewable.HasHiddenChildren[usecase.HiddenChildrenRootKey],
	})
}

// kbCreatePageRequest はページ作成の入力。
//
// parentId は任意。省略するとスペース直下（ルート）に作る。どちらで判断するかが変わる:
// 親を指定したときは「その親ページの編集権限」、省略したときは「そのスペースの編集権限」。
// ページ付与（page_grants）は経路の上から降りてくるので、親を持つ作成をスペースの
// 判定で通してはいけない（親に足された役割を取りこぼして、書ける相手を断ってしまう）。
type kbCreatePageRequest struct {
	// ParentID が空文字（未指定）ならスペース直下に作る。
	ParentID string `json:"parentId,omitempty" example:"0198a000-0000-7000-8000-000000000003"`
	Title    string `json:"title"    binding:"required,max=200" example:"設計メモ"`
}

// Create は親ページの下に新しいページを作る（親の編集権限が要る）。
func (h *KnowledgeBasePageHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	spaceID := c.Param("spaceId")
	// 判定の入口を親の有無で分ける。親があるならページの権限（祖先のページ付与まで見る）、
	// 無いならスペースの権限（ページ付与を見ない段）。取り違えると、親に editor を
	// 張られただけの相手がその下に書けない／スペースの editor がルートを作れない、
	// のどちらかになる。
	var parentID *string
	if req.ParentID == "" {
		if !h.requireSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
			return
		}
	} else {
		if !h.requirePagePermission(c, scope, req.ParentID, domain.CapabilityEdit) {
			return
		}
		parentID = &req.ParentID
	}
	page, err := h.create.Execute(c.Request.Context(), usecase.CreatePageInput{
		WorkspaceID:     scope.workspaceID,
		SpaceID:         spaceID,
		ParentID:        parentID,
		Title:           req.Title,
		CreatedByUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbPageResponse(page))
}

// Get はページ 1 件と本文を返す（閲覧権限が要る）。
func (h *KnowledgeBasePageHandler) Get(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityView) {
		return
	}
	out, err := h.get.Execute(c.Request.Context(), usecase.GetPageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 本文中のページ参照の題名を、読み手にとっての「いまの題名」へ差し替えて出す
	// （題名の正本は pages.title で、保存側は title を持たない）。解決の失敗は
	// 本文の読み出しを止めない — 元の doc のまま返し、死んでいることだけ記録する。
	doc, refErr := h.resolveRefs.Execute(c.Request.Context(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: scope.workspaceID,
		UserID:      scope.userID,
		Doc:         out.Doc,
	})
	if refErr != nil {
		slog.WarnContext(c.Request.Context(), "kb: page ref title resolve failed", "err", refErr)
	}
	c.JSON(http.StatusOK, kbPageDocResponse{
		Page: toKbPageResponse(&out.Page),
		Doc:  json.RawMessage(doc),
	})
}

type kbRenamePageRequest struct {
	Title string `json:"title" binding:"required,max=200" example:"設計メモ (改訂)"`
}

// Rename はページのタイトルを変える（編集権限が要る）。
func (h *KnowledgeBasePageHandler) Rename(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbRenamePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	page, err := h.rename.Execute(c.Request.Context(), usecase.RenamePageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		Title:       req.Title,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageResponse(page))
}

// kbMovePageRequest はページ移動の入力。
//
// parentId が必須なのは作成と同じ理由。スペース直下へ移す操作は移動先スペースに対する
// 権限で判断する必要があり、その口がまだ無い。
type kbMovePageRequest struct {
	// ParentID を省くと、いまと同じスペースの直下（ルート）へ移す。
	//
	// 省けるようにしたのはドラッグのため。入れ子になったページを最上段へ戻すのは
	// 基本の操作で、これが無いと「入れることはできるが出せない」ドラッグになる。
	// 判断はスペースの編集権限で行う（ページ付与が届かない段なので、そこが正しい単位）。
	ParentID string `json:"parentId,omitempty" example:"0198a000-0000-7000-8000-000000000003"`
	// AfterPageID / BeforePageID は移動先の兄弟の中でどこに置くかを、隣のページの ID で表す。
	// どちらも空なら末尾。**両方を指定することはできない。**
	//
	// 並び順のキーそのものを受け取らないのは、そもそも返していないため
	// （キーの整数部は兄弟の通し番号になるので、飛びから伏せた枚数が読める）。
	// 「先頭に置く」は「最初の兄弟の手前（beforePageId）」として表す。
	AfterPageID  string `json:"afterPageId,omitempty"  example:"0198a000-0000-7000-8000-000000000004"`
	BeforePageID string `json:"beforePageId,omitempty" example:"0198a000-0000-7000-8000-000000000005"`
}

// Move はページ（と子孫）を別の親の下へ移す。動かすページと移動先の親の両方に編集権限が要る。
func (h *KnowledgeBasePageHandler) Move(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	// 根の権限を先に見るのは応答を撃ち分けないため（閲覧できない根は 404 のまま）。
	// そのうえで子孫まで確かめる。
	//
	// # なぜ移動でも子孫を見るのか
	//
	// 移動はサブツリーごと動く。動いた瞬間、**子孫それぞれの祖先の並びが変わる**。
	// ページ付与（page_grants）は経路の上から降りてくるので、祖先が変われば子孫の
	// 実効権限も変わる — 移動先の祖先に張られた付与が新たに届いたり、元の祖先から
	// 届いていた付与が外れたりする。操作者はその子孫を見られないので、**自分が誰に何を
	// 開いたのか分からないまま権限を書き換えることになる。**
	// 権限を変える操作は必ず admin の gate（kb_permission_gate.go）を通すのに、
	// 移動だけがその外側から同じ結果を作れてしまう、というのがこの穴の正体。
	//
	// # なぜアーカイブと同じ判定に揃えたのか（移動特有の事情を検討したうえで）
	//
	// 「全部できるか、何もしないか」の性質が移動でもそのまま成り立つ。移動は
	// pages / page_paths / 子孫の space_id を repository の 1 トランザクションで
	// まとめて付け替える操作で、**部分的な移動という中間状態が存在しない**
	// （子孫を置き去りにすれば木が根から切れる）。アーカイブを閉じる側へ倒した論拠が
	// そのまま使えるので、判定を分ける理由が無い。分ければ「アーカイブなら断られるのに
	// 移動なら通る」という、経路で食い違う状態を自分から作ることになる。
	//
	// 移動はアーカイブより頻度が高い（ドラッグで動かせるようになれば特に）。これは
	// 緩める理由ではなく締める理由になる — 穴を踏む回数がそのまま増えるため。
	// 費用も同じで、増えるのはアーカイブと同一のクエリ 1 回だけ（実測: 5,000 ページの
	// サブツリーで 3.0 ms、最悪ケースでも 174 ms）。
	//
	// 断ること自体が「この下に触れないページがある」という粒度の粗い信号になる点も
	// アーカイブと同じで、同じ理由で許容する（どのページかまでは分からない）。
	//
	// # 断ったときに何も書き換わらないこと
	//
	// この検査は読み取りだけで、通らなければ**移動の usecase を呼ばずに return する**。
	// parent_id / position / page_paths を触るのは repository.MovePage だけなので、
	// ここで返した時点でどのテーブルにも書き込みは起きていない。
	//
	// # 同一スペース内の移動もここを通る
	//
	// repository.ErrPageMoveVoidsSpaceGrant が塞いでいるのはスペースをまたぐ
	// 移動だけ（「スペース全員」宛てのページ付与が移動先で失効する場合）。同一スペース内で
	// 親を付け替える移動には、子孫の権限を見る経路がこれ以外に無い。
	if !h.requireSubtreeEditPermission(c, scope, pageID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbMovePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// 移動先も編集できなければならない。動かすページの権限だけで通すと、
	// 自分が書けないサブツリーへページを差し込めてしまう。
	//
	// 親を指定したときは**その親ページ**の編集権限、省いたとき（スペース直下へ戻す）は
	// **そのスペース**の編集権限で判断する。作成の入口と同じ分け方で、理由も同じ —
	// ページ付与（page_grants）は経路の上から降りてくるので、親を持つ移動を
	// スペースの判定で通すと、親に張られた付与を取りこぼして書ける相手を断ってしまう。
	// 逆にスペース直下にはページ付与が届かないので、そこはスペースの権限が正しい単位。
	var newParentID *string
	if req.ParentID != "" {
		if !h.requirePagePermission(c, scope, req.ParentID, domain.CapabilityEdit) {
			return
		}
		newParentID = &req.ParentID
	} else {
		// 動かすページ自身の所属スペースへ戻す（スペースをまたぐ移動はこの口では扱わない）。
		// ページの編集権限は上で確かめてあるので、ここで読んでも実在は新しく漏れない。
		moving, err := h.findPage.Execute(c.Request.Context(), usecase.FindPageInput{
			WorkspaceID: scope.workspaceID,
			PageID:      pageID,
		})
		if err != nil {
			respondKnowledgeBaseErr(c, err)
			return
		}
		if !h.requireSpacePermission(c, scope, moving.SpaceID, domain.CapabilityEdit) {
			return
		}
	}
	// 位置の指定は前後どちらか一方だけ。両方あると、どちらを採ったかで結果が変わるのに
	// 呼び出し側からは分からない。黙って片方を採らず、断る。
	if req.AfterPageID != "" && req.BeforePageID != "" {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	anchor := req.AfterPageID
	anchorBefore := false
	if req.BeforePageID != "" {
		anchor = req.BeforePageID
		anchorBefore = true
	}
	// 隣に指定したページは**閲覧できなければならない**。
	//
	// 確かめずに通すと、「その ID が移動先の子か」を成功と 400 の差で言い当てられる
	// （移動先を編集できれば誰でも叩ける）。閲覧できるページなら、そこに在ることは
	// 既に分かっているので新しくは漏れない。閲覧できなければ他と同じ 404 に畳まれる。
	if anchor != "" && !h.requirePagePermission(c, scope, anchor, domain.CapabilityView) {
		return
	}
	page, err := h.move.Execute(c.Request.Context(), usecase.MovePageInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		NewParentID:  newParentID,
		Anchor:       anchor,
		AnchorBefore: anchorBefore,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageResponse(page))
}

// Archive はページと子孫をまとめてアーカイブする（編集権限が要る）。
func (h *KnowledgeBasePageHandler) Archive(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	// 根の権限を先に見るのは応答を撃ち分けないため（閲覧できない根は 404 のまま）。
	// そのうえで子孫まで確かめる。
	if !h.requireSubtreeEditPermission(c, scope, pageID) {
		return
	}
	if err := h.archive.Execute(c.Request.Context(), usecase.ArchivePageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	}); err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete はページを子孫ごと物理削除する（戻せない）。
func (h *KnowledgeBasePageHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	// アーカイブと同じ入口: 根の権限を先に見て応答を撃ち分けず、そのうえで子孫まで確かめる。
	//
	// 既知の限界（アーカイブ・移動も同じ形）: 検査と DELETE は別々のクエリで、
	// 間に別ユーザーの移動が挟まると「検査していないページ」を CASCADE が道連れに
	// し得る。窓はミリ秒で、成立には同一ワークスペースの編集者どうしの同時操作が要る。
	// 塞ぐには検査と削除を同一トランザクションで行ロックする必要があり、
	// 権限の事実集めが別リポジトリにある現構成では境界の作り直しになるため、
	// 直列化はその再設計（権限操作の口の統合）とセットで行う。
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	if !h.requireSubtreeEditPermission(c, scope, pageID) {
		return
	}
	if err := h.deletePage.Execute(c.Request.Context(), usecase.DeletePageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	}); err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Unarchive はアーカイブしたページを現役へ戻す（編集権限が要る）。
func (h *KnowledgeBasePageHandler) Unarchive(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	if !h.requireSubtreeEditPermission(c, scope, pageID) {
		return
	}
	page, err := h.unarchive.Execute(c.Request.Context(), usecase.UnarchivePageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageResponse(page))
}

type kbReplaceContentRequest struct {
	Doc json.RawMessage `json:"doc" binding:"required"`
}

// ReplaceContent はページ本文を丸ごと置き換える（編集権限が要る）。
func (h *KnowledgeBasePageHandler) ReplaceContent(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePagePermission(c, scope, pageID, domain.CapabilityEdit) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbReplaceContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	snap, err := h.replaceBlocks.Execute(c.Request.Context(), usecase.ReplacePageBlocksInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		Doc:         string(req.Doc),
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, kbPageContentResponse{Doc: json.RawMessage(snap.Doc), BuiltAt: snap.BuiltAt})
}

// kbPageContentResponse は本文置き換えの結果（保存された正規形と、その焼き直し時刻）。
type kbPageContentResponse struct {
	Doc     json.RawMessage `json:"doc"`
	BuiltAt time.Time       `json:"builtAt"`
}

// limitKnowledgeBaseBody は bind 前にボディサイズ上限を課す。
func limitKnowledgeBaseBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxKnowledgeBaseBodyBytes)
}

// kbResolvedPageResponse は /kb/pages/{pageId}（URL にテナントを持たない解決）の返却形。
// workspaceSlug は以降の API 呼び出し（木・保存）に、workspaceName と ancestors は
// パンくず（場所の表示）に、canEdit は編集 UI の、canManage は共有 UI の出し分けに使う。
// ancestors は**読み手が閲覧できる祖先だけ**を根から順に持つ（木と同じ規則で穴があき得る）。
type kbResolvedPageResponse struct {
	WorkspaceSlug string          `json:"workspaceSlug" example:"w-3f2a9c"`
	WorkspaceName string          `json:"workspaceName" example:"開発チーム"`
	Page          kbPageResponse  `json:"page"`
	Doc           json.RawMessage `json:"doc"`
	CanEdit       bool            `json:"canEdit"`
	// CanManage はそのページの権限を変えられるか（共有ボタンを出すかの判定に使う）。
	// 届いている役割が admin かどうかだけで決まる。
	CanManage bool                  `json:"canManage"`
	Ancestors []usecase.AncestorRef `json:"ancestors"`
}

// ResolveByID は /p/{pageId} の URL からページを開く（URL にワークスペースを出さないための口）。
func (h *KnowledgeBasePageHandler) ResolveByID(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	pageID := c.Param("pageId")
	loc, err := h.resolve.Execute(c.Request.Context(), pageID)
	if err != nil {
		// 実在しない ID も、この後の権限で伏せられる ID も、同じ経路の 404 に落ちる。
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 解決はテナント確定前の読みなので、**ここで必ず**その workspace の権限判定を通す。
	perm, err := h.check.Execute(c.Request.Context(), usecase.CheckPagePermissionInput{
		WorkspaceID: loc.Workspace.ID,
		PageID:      pageID,
		UserID:      uid,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	if !perm.CanView {
		// 閲覧できない相手にはページの実在を教えない（存在しない ID と同じ応答）。
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return
	}
	out, err := h.get.Execute(c.Request.Context(), usecase.GetPageInput{
		WorkspaceID: loc.Workspace.ID,
		PageID:      pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// Get と同じく、本文中のページ参照の題名を読み手の可視範囲で解決して出す。
	doc, refErr := h.resolveRefs.Execute(c.Request.Context(), usecase.ResolvePageRefTitlesInput{
		WorkspaceID: loc.Workspace.ID,
		UserID:      uid,
		Doc:         out.Doc,
	})
	if refErr != nil {
		slog.WarnContext(c.Request.Context(), "kb: page ref title resolve failed", "err", refErr)
	}
	// パンくず（閲覧できる祖先だけ）。失敗してもページは開く — 空のまま出し、記録だけ残す。
	ancestors, ancErr := h.ancestors.Execute(c.Request.Context(), usecase.ListViewableAncestorsInput{
		WorkspaceID: loc.Workspace.ID,
		UserID:      uid,
		PageID:      pageID,
	})
	if ancErr != nil {
		slog.WarnContext(c.Request.Context(), "kb: ancestors resolve failed", "err", ancErr)
		ancestors = []usecase.AncestorRef{}
	}
	c.JSON(http.StatusOK, kbResolvedPageResponse{
		WorkspaceSlug: loc.Workspace.Slug,
		WorkspaceName: loc.Workspace.Name,
		Page:          toKbPageResponse(&out.Page),
		Doc:           json.RawMessage(doc),
		CanEdit:       perm.CanEdit,
		CanManage:     perm.CanManage,
		Ancestors:     ancestors,
	})
}
