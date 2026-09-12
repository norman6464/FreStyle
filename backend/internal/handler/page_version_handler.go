package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// PageVersionHandler はページ本文の版（page_versions）を受ける。CommentHandler と同じ形——
// requirePagePermissionWith を再利用し、user.LookupUserDisplayUseCase で著者を解決する。
// バージョンは CanComment のような新しい権限軸を持たず、ページ本文そのものの履歴なので
// 認可は CapabilityView / CapabilityEdit の 2 値だけで足りる（一覧・単体取得は閲覧できれば
// 誰でも、作成・復元は編集できる人だけ）。
type PageVersionHandler struct {
	check       *kb.CheckPagePermissionUseCase
	create      *kb.CreateExplicitPageVersionUseCase
	list        *kb.ListPageVersionsUseCase
	get         *kb.GetPageVersionUseCase
	restore     *kb.RestorePageVersionUseCase
	userDisplay *user.LookupUserDisplayUseCase
}

func NewPageVersionHandler(
	check *kb.CheckPagePermissionUseCase,
	create *kb.CreateExplicitPageVersionUseCase,
	list *kb.ListPageVersionsUseCase,
	get *kb.GetPageVersionUseCase,
	restore *kb.RestorePageVersionUseCase,
	userDisplay *user.LookupUserDisplayUseCase,
) *PageVersionHandler {
	return &PageVersionHandler{check: check, create: create, list: list, get: get, restore: restore, userDisplay: userDisplay}
}

// pageVersionSummaryResponse は一覧の 1 要素。doc は含まない（一覧はメタ情報だけで十分で、
// ページの版が多いほど doc を毎回積むと応答が重くなるため）。
type pageVersionSummaryResponse struct {
	Seq       int64               `json:"seq"`
	Author    userDisplayResponse `json:"author"`
	Note      *string             `json:"note,omitempty"`
	CreatedAt time.Time           `json:"createdAt"`
}

func (h *PageVersionHandler) toSummaryResponse(ctx context.Context, v domain.PageVersion, cache userDisplayCache) pageVersionSummaryResponse {
	return pageVersionSummaryResponse{
		Seq:       v.Seq,
		Author:    resolveUserDisplay(ctx, h.userDisplay, v.AuthorUserID, cache),
		Note:      v.Note,
		CreatedAt: v.CreatedAt,
	}
}

// pageVersionDetailResponse は単体取得・作成の応答（一覧の形 + doc）。
type pageVersionDetailResponse struct {
	Seq       int64               `json:"seq"`
	Author    userDisplayResponse `json:"author"`
	Note      *string             `json:"note,omitempty"`
	CreatedAt time.Time           `json:"createdAt"`
	// Doc は ProseMirror ドキュメント（tiptap の getJSON() 相当）。
	Doc json.RawMessage `json:"doc"`
}

func (h *PageVersionHandler) toDetailResponse(ctx context.Context, v domain.PageVersion, cache userDisplayCache) pageVersionDetailResponse {
	return pageVersionDetailResponse{
		Seq:       v.Seq,
		Author:    resolveUserDisplay(ctx, h.userDisplay, v.AuthorUserID, cache),
		Note:      v.Note,
		CreatedAt: v.CreatedAt,
		Doc:       json.RawMessage(v.Doc),
	}
}

// List はページの版一覧を返す（CapabilityView — 閲覧できれば誰でも読める。doc は含まない）。
func (h *PageVersionHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityView) {
		return
	}
	out, err := h.list.Execute(c.Request.Context(), kb.ListPageVersionsInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	versions := make([]pageVersionSummaryResponse, 0, len(out))
	for _, v := range out {
		versions = append(versions, h.toSummaryResponse(c.Request.Context(), v, cache))
	}
	c.JSON(http.StatusOK, versions)
}

// parsePageVersionSeq はパス変数 :seq を int64 としてパースする。パース失敗は
// domain.ErrPageVersionNotFound と同じ扱い（404）にする（実在確認を漏らさないため）。
func parsePageVersionSeq(c *gin.Context) (int64, bool) {
	seq, err := strconv.ParseInt(c.Param("seq"), 10, 64)
	if err != nil {
		respondKnowledgeBaseErr(c, domain.ErrPageVersionNotFound)
		return 0, false
	}
	return seq, true
}

// Get は版 1 件（doc 込み）を返す（CapabilityView）。
func (h *PageVersionHandler) Get(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityView) {
		return
	}
	seq, ok := parsePageVersionSeq(c)
	if !ok {
		return
	}
	v, err := h.get.Execute(c.Request.Context(), kb.GetPageVersionInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		Seq:         seq,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusOK, h.toDetailResponse(c.Request.Context(), *v, cache))
}

// kbCreatePageVersionRequest は「版を残す」の入力。note は任意（空文字・空白のみ・未指定は
// domain.ValidateVersionNote が nil へ正規化する）。
type kbCreatePageVersionRequest struct {
	Note *string `json:"note,omitempty"`
}

// Create は「版を残す」— 10 分規則を無視して必ず 1 件版を切る（CapabilityEdit）。
func (h *PageVersionHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityEdit) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreatePageVersionRequest
	// body 無し（EOF、note を送らない呼び出し）は許容し、壊れた JSON だけ 400 で弾く。
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	v, err := h.create.Execute(c.Request.Context(), kb.CreateExplicitPageVersionInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		AuthorUserID: scope.userID,
		Note:         req.Note,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusCreated, h.toDetailResponse(c.Request.Context(), *v, cache))
}

// Restore は過去の版を今の本文として復元する（CapabilityEdit）。応答は ReplaceContent
// （PUT .../content）と同じ形に揃える（kbPageContentResponse を再利用）。
func (h *PageVersionHandler) Restore(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityEdit) {
		return
	}
	seq, ok := parsePageVersionSeq(c)
	if !ok {
		return
	}
	snap, err := h.restore.Execute(c.Request.Context(), kb.RestorePageVersionInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		Seq:          seq,
		EditorUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 復元した本人が最終編集者になる。
	editorID := scope.userID
	builtAt := snap.BuiltAt
	c.JSON(http.StatusOK, kbPageContentResponse{
		Doc:          json.RawMessage(snap.Doc),
		BuiltAt:      snap.BuiltAt,
		LastEditedBy: kbLastEditedByResponseWith(c.Request.Context(), h.userDisplay, &editorID),
		LastEditedAt: &builtAt,
	})
}
