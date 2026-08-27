package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// KnowledgeBasePageHandler はナレッジ基盤のページ操作を受ける。
//
// ワークスペースはリクエストからは受け取らず、middleware.KnowledgeBaseWorkspace が
// URL の slug と principals から確定させたものを context から取る。
type KnowledgeBasePageHandler struct {
	check          *usecase.CheckPagePermissionUseCase
	checkSpace     *usecase.CheckSpacePermissionUseCase
	canEditSubtree *usecase.CanEditPageSubtreeUseCase
	listViewable   *usecase.ListViewablePagesUseCase
	get            *usecase.GetPageUseCase
	create         *usecase.CreatePageUseCase
	rename         *usecase.RenamePageUseCase
	move           *usecase.MovePageUseCase
	archive        *usecase.ArchivePageUseCase
	unarchive      *usecase.UnarchivePageUseCase
	replaceBlocks  *usecase.ReplacePageBlocksUseCase
}

// NewKnowledgeBasePageHandler は KnowledgeBasePageHandler を組み立てる。
func NewKnowledgeBasePageHandler(
	check *usecase.CheckPagePermissionUseCase,
	checkSpace *usecase.CheckSpacePermissionUseCase,
	canEditSubtree *usecase.CanEditPageSubtreeUseCase,
	listViewable *usecase.ListViewablePagesUseCase,
	get *usecase.GetPageUseCase,
	create *usecase.CreatePageUseCase,
	rename *usecase.RenamePageUseCase,
	move *usecase.MovePageUseCase,
	archive *usecase.ArchivePageUseCase,
	unarchive *usecase.UnarchivePageUseCase,
	replaceBlocks *usecase.ReplacePageBlocksUseCase,
) *KnowledgeBasePageHandler {
	return &KnowledgeBasePageHandler{
		check:          check,
		checkSpace:     checkSpace,
		canEditSubtree: canEditSubtree,
		listViewable:   listViewable,
		get:            get,
		create:         create,
		rename:         rename,
		move:           move,
		archive:        archive,
		unarchive:      unarchive,
		replaceBlocks:  replaceBlocks,
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

func toKbPageTreeResponse(nodes []*usecase.PageTreeNode, hidden map[string]bool) []kbPageTreeResponse {
	out := make([]kbPageTreeResponse, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, kbPageTreeResponse{
			Page:              toKbPageResponse(&n.Page),
			Children:          toKbPageTreeResponse(n.Children, hidden),
			HasHiddenChildren: hidden[n.Page.ID],
		})
	}
	return out
}

// kbPageDocResponse はページのメタ情報と本文（ProseMirror doc）の組。
type kbPageDocResponse struct {
	Page kbPageResponse  `json:"page"`
	Doc  json.RawMessage `json:"doc" swaggertype:"object"`
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
	case errors.Is(err, repository.ErrPageMoveVoidsSpaceRestriction):
		// 「今の権限設定のままでは移せない」という業務上の衝突であって、サーバの故障ではない。
		// 既にアーカイブ済み・循環と同じ 409 に揃える（どれも「リクエスト自体は正しいが、
		// 対象の現在の状態と両立しない」）。500 で返すと、クライアントは DB 障害と区別できず
		// 再試行してよいものと誤解する（何度試しても同じ結果になる）。
		c.JSON(http.StatusConflict, errorResponse{Error: "space_restriction_voided"})
	case errors.Is(err, repository.ErrWorkspaceSlugTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "slug_taken"})
	case errors.Is(err, repository.ErrSpaceKeyTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "space_key_taken"})
	case errors.Is(err, usecase.ErrInvalidWorkspaceSlug),
		errors.Is(err, usecase.ErrInvalidSpaceKey),
		errors.Is(err, usecase.ErrInvalidName):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
	case errors.Is(err, usecase.ErrPageParentSpaceMismatch):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "parent_space_mismatch"})
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
// **ページを名指しする経路でこれを使ってはいけない。** スペースの判定はページ単位の例外
// （page_restrictions）を見ておらず、あるページで deny されている相手にも
// スペースの既定が editor なら true を返す。使ってよいのは対象がまだ存在しない操作
// （スペース直下へのページ作成）だけで、親を持つ作成は requirePagePermission を通す。
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
// 根 1 枚だけを見ないのは、同じ「編集」の判定が経路で食い違わないようにするため。
// 子孫には親と違う例外を張れるので、根だけで通すと、直接 rename すれば 403 になる子を
// 祖先のアーカイブ経由で書き換えられる（管理者のツリーからも消える）。部分的に
// アーカイブして逃げる手も採れない — アーカイブ済みの親の下に現役の子が残ると
// ツリーに現れない迷子ページになり、復帰の前提（親から順に戻す）も壊れる。
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
//
//	@Summary      ナレッジ 基盤 の ページ ツリー
//	@Description  スペース 配下 の 現役 ページ の うち 閲覧 できる もの だけ を 木 で 返す。 見え ない 親 の 配下 は (権限 が あっ て も) ツリー に は 現れ ない。 存在 し ない スペース と 中身 が 1 件 も 見え ない スペース は 区別 し ない (どちら も 空 の pages)。 hasHiddenChildren は その 段 の 直下 に 閲覧 でき ない ページ が 在る か で、 枚数 も 題名 も 返さ ない。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        spaceId        path      string  true  "スペース ID (UUID)"
//	@Success      200            {object}  kbPageTreeRootResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "ワークスペース が 無い か 未 所属"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/pages [get]
//	@Security     CookieAuth
func (h *KnowledgeBasePageHandler) Tree(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	// ページごとに権限を引くと N+1 になるので、一覧はまとめて 1 回で解決する。
	viewable, err := h.listViewable.Execute(c.Request.Context(), usecase.ListViewablePagesInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     c.Param("spaceId"),
		UserID:      scope.userID,
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
	tree := usecase.BuildPageTree(viewable.Pages, usecase.PageTreeOrphanHidden)
	c.JSON(http.StatusOK, kbPageTreeRootResponse{
		Pages:             toKbPageTreeResponse(tree, viewable.HasHiddenChildren),
		HasHiddenChildren: viewable.HasHiddenChildren[usecase.HiddenChildrenRootKey],
	})
}

// kbCreatePageRequest はページ作成の入力。
//
// parentId は任意。省略するとスペース直下（ルート）に作る。どちらで判断するかが変わる:
// 親を指定したときは「その親ページの編集権限」、省略したときは「そのスペースの編集権限」。
// ページの例外（page_restrictions）は経路の上から効くので、親を持つ作成をスペースの
// 判定で通してはいけない（親で deny されている相手がその下に書けてしまう）。
type kbCreatePageRequest struct {
	// ParentID が空文字（未指定）ならスペース直下に作る。
	ParentID string `json:"parentId,omitempty" example:"0198a000-0000-7000-8000-000000000003"`
	Title    string `json:"title"    binding:"required,max=200" example:"設計メモ"`
}

// Create は親ページの下に新しいページを作る（親の編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ 作成
//	@Description  parentId の 下 に ページ を 作る。 親 を 編集 できる 者 だけ が 作れる。 親 が 閲覧 でき ない 場合 は 存在 を 漏らさ ず 404。 parentId を 省略 する と スペース 直下 (ルート) に 作り、 この とき は スペース の 編集 権限 で 判断 する (スペース に は ページ 単位 の 例外 が 無い ため。 親 を 指定 し た 作成 は 必ず 親 ページ の 権限 で 判断 する)。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string               true  "ワークスペース の slug"
//	@Param        spaceId        path      string               true  "スペース ID (UUID)"
//	@Param        body           body      kbCreatePageRequest  true  "作成 内容 (parentId/title 必須)"
//	@Success      201            {object}  kbPageResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "親 を 編集 する 権限 が 無い"
//	@Failure      404            {object}  errorResponse  "スペース / 親 が 無い か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "親 が アーカイブ 済み"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/pages [post]
//	@Security     CookieAuth
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
	// 判定の入口を親の有無で分ける。親があるならページの権限（経路上の例外まで見る）、
	// 無いならスペースの権限（例外の層が無い）。取り違えると、親で deny されている相手が
	// その下にページを足せる／スペースの editor がルートを作れない、のどちらかになる。
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
//
//	@Summary      ナレッジ 基盤 の ページ 取得
//	@Description  ページ の メタ 情報 と 本文 (ProseMirror doc) を 返す。 閲覧 権限 が 無い ページ は 存在 し ない ページ と 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        pageId         path      string  true  "ページ ID (UUID)"
//	@Success      200            {object}  kbPageDocResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId} [get]
//	@Security     CookieAuth
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
	c.JSON(http.StatusOK, kbPageDocResponse{
		Page: toKbPageResponse(&out.Page),
		Doc:  json.RawMessage(out.Doc),
	})
}

type kbRenamePageRequest struct {
	Title string `json:"title" binding:"required,max=200" example:"設計メモ (改訂)"`
}

// Rename はページのタイトルを変える（編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ 改名
//	@Description  タイトル だけ を 変更 する。 編集 権限 が 要る。 アーカイブ 済み ページ は 変更 でき ない (409)。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string               true  "ワークスペース の slug"
//	@Param        pageId         path      string               true  "ページ ID (UUID)"
//	@Param        body           body      kbRenamePageRequest  true  "新しい タイトル"
//	@Success      200            {object}  kbPageResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "アーカイブ 済み"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId} [patch]
//	@Security     CookieAuth
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
	ParentID string `json:"parentId" binding:"required" example:"0198a000-0000-7000-8000-000000000003"`
}

// Move はページ（と子孫）を別の親の下へ移す。動かすページと移動先の親の両方に編集権限が要る。
//
//	@Summary      ナレッジ 基盤 の ページ 移動
//	@Description  ページ を parentId の 下 へ 移す。 動かす ページ と 移動 先 の 親 の 両方 に 編集 権限 が 要る (片方 だけ で 移せる と 書け ない 場所 へ 書き込め て しまう)。 さらに 動かす ページ の 子孫 すべて に 編集 権限 が 要る (1 枚 でも 編集 でき ない ページ が 配下 に あれ ば 403 subtree_forbidden で 何 も 書き換え ない)。 移動 は サブツリー ごと 動く の で、 操作 者 から 見え ない 子孫 の 祖先 まで 変わり、 そこ から 継承 さ れる 権限 が 本人 の 知ら ない うち に 変わる ため。 アーカイブ / 復帰 と 同じ 判定 に 揃え て ある。 スペース 直下 へ の 移動 は 未 対応。 動かす サブツリー に 「スペース 全員」 宛て の 例外 が 残っ て いる 状態 で 別 スペース へ 移す 操作 は 409 (space_restriction_voided) で 断る。 例外 を 先 に 整理 し て から 移す。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string             true  "ワークスペース の slug"
//	@Param        pageId         path      string             true  "ページ ID (UUID)"
//	@Param        body           body      kbMovePageRequest  true  "移動 先 の 親"
//	@Success      200            {object}  kbPageResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー / スペース 不一致"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い / 配下 に 編集 でき ない ページ が ある"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "アーカイブ 済み / 循環 / スペース 全員 宛て の 例外 が 失効 する 移動"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/move [post]
//	@Security     CookieAuth
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
	// ページの例外（page_restrictions / page_allow_lists）は経路の上から効くので、
	// 祖先が変われば子孫の実効権限も変わる — 限定公開だったページが移動先の既定で
	// 開いたり、逆に見えていた相手から消えたりする。操作者はその子孫を見られないので、
	// **自分が何を open / close したのか分からないまま権限を書き換えることになる。**
	// 権限を変える操作は例外なく admin の gate（kb_permission_gate.go）を通すのに、
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
	// repository.ErrPageMoveVoidsSpaceRestriction が塞いでいるのはスペースをまたぐ
	// 移動だけ（「スペース全員」宛ての例外が移動先で失効する場合）。同一スペース内で
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
	// 移動先の親も編集できなければならない。動かすページの権限だけで通すと、
	// 自分が書けないサブツリーへページを差し込めてしまう。
	if !h.requirePagePermission(c, scope, req.ParentID, domain.CapabilityEdit) {
		return
	}
	page, err := h.move.Execute(c.Request.Context(), usecase.MovePageInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		NewParentID: &req.ParentID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageResponse(page))
}

// Archive はページと子孫をまとめてアーカイブする（編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ アーカイブ
//	@Description  ページ と その 子孫 を まとめて ツリー から 隠す。 対象 の ページ だけ で なく 子孫 すべて に 編集 権限 が 要る (1 枚 でも 編集 でき ない ページ が 配下 に あれ ば 403 subtree_forbidden で 何 も し ない)。 これ は 意図 し た 設計 で、 同じ ページ を 直接 改名 する 場合 と 判定 を 揃える ため。 既に アーカイブ 済み なら 何 も し ない (冪等)。
//	@Tags         knowledge-base
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Success      204            "アーカイブ 成功 (本文 なし)"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い / 配下 に 編集 でき ない ページ が ある"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/archive [post]
//	@Security     CookieAuth
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

// Unarchive はアーカイブしたページを現役へ戻す（編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ 復帰
//	@Description  アーカイブ した ページ を (同時 に アーカイブ さ れ た 子孫 ごと) 現役 へ 戻す。 アーカイブ と 同じ く 子孫 すべて に 編集 権限 が 要る (1 枚 でも 編集 でき ない ページ が 配下 に あれ ば 403 subtree_forbidden)。 親 が まだ アーカイブ 中 なら 戻せ ない (409)。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        pageId         path      string  true  "ページ ID (UUID)"
//	@Success      200            {object}  kbPageResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い / 配下 に 編集 でき ない ページ が ある"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "親 が アーカイブ 中"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/unarchive [post]
//	@Security     CookieAuth
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
	Doc json.RawMessage `json:"doc" binding:"required" swaggertype:"object"`
}

// ReplaceContent はページ本文を丸ごと置き換える（編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ 本文 置き換え
//	@Description  ページ の 本文 (ProseMirror doc) を 丸ごと 置き換える。 編集 権限 が 要る。 保存 さ れる の は 行 スキーマ から 組み立て 直し た 正規 形 で、 レスポンス は その 正規 形 を 返す。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                   true  "ワークスペース の slug"
//	@Param        pageId         path      string                   true  "ページ ID (UUID)"
//	@Param        body           body      kbReplaceContentRequest  true  "本文 (ProseMirror doc)"
//	@Success      200            {object}  kbPageContentResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー / doc が 壊れ て いる"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "アーカイブ 済み"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/content [put]
//	@Security     CookieAuth
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
	Doc     json.RawMessage `json:"doc" swaggertype:"object"`
	BuiltAt time.Time       `json:"builtAt"`
}

// limitKnowledgeBaseBody は bind 前にボディサイズ上限を課す。
func limitKnowledgeBaseBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxKnowledgeBaseBodyBytes)
}
