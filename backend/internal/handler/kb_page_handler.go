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
	check         *usecase.CheckPagePermissionUseCase
	listViewable  *usecase.ListViewablePagesUseCase
	get           *usecase.GetPageUseCase
	create        *usecase.CreatePageUseCase
	rename        *usecase.RenamePageUseCase
	move          *usecase.MovePageUseCase
	archive       *usecase.ArchivePageUseCase
	unarchive     *usecase.UnarchivePageUseCase
	replaceBlocks *usecase.ReplacePageBlocksUseCase
}

// NewKnowledgeBasePageHandler は KnowledgeBasePageHandler を組み立てる。
func NewKnowledgeBasePageHandler(
	check *usecase.CheckPagePermissionUseCase,
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
		check:         check,
		listViewable:  listViewable,
		get:           get,
		create:        create,
		rename:        rename,
		move:          move,
		archive:       archive,
		unarchive:     unarchive,
		replaceBlocks: replaceBlocks,
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
type kbPageResponse struct {
	ID              string     `json:"id"              example:"0198a000-0000-7000-8000-000000000003"`
	SpaceID         string     `json:"spaceId"         example:"0198a000-0000-7000-8000-000000000002"`
	ParentID        *string    `json:"parentId,omitempty"`
	Position        string     `json:"position"        example:"a0"`
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
		Position:        p.Position,
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
}

func toKbPageTreeResponse(nodes []*usecase.PageTreeNode) []kbPageTreeResponse {
	out := make([]kbPageTreeResponse, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, kbPageTreeResponse{
			Page:     toKbPageResponse(&n.Page),
			Children: toKbPageTreeResponse(n.Children),
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

// Tree はスペース配下の、そのユーザーが閲覧できるページを木構造で返す。
//
//	@Summary      ナレッジ 基盤 の ページ ツリー
//	@Description  スペース 配下 の 現役 ページ の うち 閲覧 できる もの だけ を 木 で 返す。 見え ない 親 の 配下 は (権限 が あっ て も) ツリー に は 現れ ない。 存在 し ない スペース と 中身 が 1 件 も 見え ない スペース は 区別 し ない (どちら も 空 配列)。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        spaceId        path      string  true  "スペース ID (UUID)"
//	@Success      200            {array}   kbPageTreeResponse
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
	pages, err := h.listViewable.Execute(c.Request.Context(), usecase.ListViewablePagesInput{
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
	c.JSON(http.StatusOK, toKbPageTreeResponse(usecase.BuildPageTree(pages, usecase.PageTreeOrphanHidden)))
}

// kbCreatePageRequest はページ作成の入力。
//
// parentId は必須。スペース直下（親なし）への作成は、ページではなくスペースに対する
// 編集権限で判断する必要があるが、権限モデルはまだスペース単位の実効権限を返す口を持たない。
// 「メンバーなら誰でも root ページを作れる」という緩い実装で埋めると、あとから締めるのが
// 難しい穴になるため、確実に判断できる「親ページの編集権限」だけを入口にしている。
type kbCreatePageRequest struct {
	ParentID string `json:"parentId" binding:"required" example:"0198a000-0000-7000-8000-000000000003"`
	Title    string `json:"title"    binding:"required,max=200" example:"設計メモ"`
}

// Create は親ページの下に新しいページを作る（親の編集権限が要る）。
//
//	@Summary      ナレッジ 基盤 の ページ 作成
//	@Description  parentId の 下 に ページ を 作る。 親 を 編集 できる 者 だけ が 作れる。 親 が 閲覧 でき ない 場合 は 存在 を 漏らさ ず 404。 スペース 直下 へ の 作成 は 未 対応 (parentId は 必須)。
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
	if !h.requirePagePermission(c, scope, req.ParentID, domain.CapabilityEdit) {
		return
	}
	page, err := h.create.Execute(c.Request.Context(), usecase.CreatePageInput{
		WorkspaceID:     scope.workspaceID,
		SpaceID:         c.Param("spaceId"),
		ParentID:        &req.ParentID,
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
//	@Description  ページ を parentId の 下 へ 移す。 動かす ページ と 移動 先 の 親 の 両方 に 編集 権限 が 要る (片方 だけ で 移せる と 書け ない 場所 へ 書き込め て しまう)。 スペース 直下 へ の 移動 は 未 対応。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string             true  "ワークスペース の slug"
//	@Param        pageId         path      string             true  "ページ ID (UUID)"
//	@Param        body           body      kbMovePageRequest  true  "移動 先 の 親"
//	@Success      200            {object}  kbPageResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー / スペース 不一致"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い"
//	@Failure      404            {object}  errorResponse  "存在 し ない か 閲覧 権限 が 無い"
//	@Failure      409            {object}  errorResponse  "アーカイブ 済み / 循環"
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
//	@Description  ページ と その 子孫 を まとめて ツリー から 隠す。 編集 権限 が 要る。 既に アーカイブ 済み なら 何 も し ない (冪等)。
//	@Tags         knowledge-base
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Success      204            "アーカイブ 成功 (本文 なし)"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い"
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
//	@Description  アーカイブ した ページ を (同時 に アーカイブ さ れ た 子孫 ごと) 現役 へ 戻す。 編集 権限 が 要る。 親 が まだ アーカイブ 中 なら 戻せ ない (409)。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        pageId         path      string  true  "ページ ID (UUID)"
//	@Success      200            {object}  kbPageResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      403            {object}  errorResponse  "編集 権限 が 無い"
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
