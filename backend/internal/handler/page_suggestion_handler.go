package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
)

// PageSuggestionHandler は commenter が保存した提案（page_suggestions）を受ける。
// page_template_handler.go / page_version_handler.go と同じ形（kbRequestScope・
// respondKnowledgeBaseErr・requirePagePermissionWith / requireCommentPermissionWith の再利用）。
//
// 認可の方針: 作成は CanComment（requireCommentPermissionWith — comment_handler.go と同じ判定）、
// 一覧の閲覧は CanView（コメント一覧が誰でも見られるのと同じ考え方）、採用・却下は CanEdit。
type PageSuggestionHandler struct {
	check      *kb.CheckPagePermissionUseCase
	create     *kb.CreateSuggestionUseCase
	listOpen   *kb.ListOpenPageSuggestionsUseCase
	accept     *kb.AcceptPageSuggestionUseCase
	reject     *kb.RejectPageSuggestionUseCase
	getVersion *kb.GetPageVersionUseCase
	userName   *kb.LookupUserNameUseCase
}

// NewPageSuggestionHandler は PageSuggestionHandler を組み立てる。
func NewPageSuggestionHandler(
	check *kb.CheckPagePermissionUseCase,
	create *kb.CreateSuggestionUseCase,
	listOpen *kb.ListOpenPageSuggestionsUseCase,
	accept *kb.AcceptPageSuggestionUseCase,
	reject *kb.RejectPageSuggestionUseCase,
	getVersion *kb.GetPageVersionUseCase,
	userName *kb.LookupUserNameUseCase,
) *PageSuggestionHandler {
	return &PageSuggestionHandler{
		check: check, create: create, listOpen: listOpen, accept: accept, reject: reject,
		getVersion: getVersion, userName: userName,
	}
}

// pageSuggestionNameCache は 1 リクエストの応答を組み立てる間だけ使うユーザー名のその場限りの
// キャッシュ（commentNameCache / pageVersionNameCache と同じ役割）。
type pageSuggestionNameCache map[uint64]string

// resolveRef はユーザー ID を著者・解決者の応答形へ解決する。名前の解決に失敗しても応答は
// 止めない（CommentHandler.resolveAuthorRef と同じ扱い。空文字で埋めてログだけ残す）。
func (h *PageSuggestionHandler) resolveRef(ctx context.Context, userID uint64, cache pageSuggestionNameCache) kbEditorRefResponse {
	name, ok := cache[userID]
	if !ok {
		var err error
		name, err = h.userName.Execute(ctx, userID)
		if err != nil {
			slog.WarnContext(ctx, "page suggestion: name resolve failed", "err", err)
		}
		cache[userID] = name
	}
	return kbEditorRefResponse{UserID: userID, Name: name}
}

// kbPageSuggestionResponse は提案 1 件の返却形。
type kbPageSuggestionResponse struct {
	ID         string               `json:"id"`
	BaseSeq    *int64               `json:"baseSeq,omitempty"`
	Doc        json.RawMessage      `json:"doc"`
	Status     string               `json:"status"`
	Author     kbEditorRefResponse  `json:"author"`
	CreatedAt  time.Time            `json:"createdAt"`
	ResolvedAt *time.Time           `json:"resolvedAt,omitempty"`
	ResolvedBy *kbEditorRefResponse `json:"resolvedBy,omitempty"`
	// BaseDoc は BaseSeq が指す版の本文（差分表示用の付随情報）。BaseSeq が nil、または
	// その版が既に引けない場合は省略する（診断情報でしかないので、それだけで提案自体の
	// 応答を止めない）。
	BaseDoc json.RawMessage `json:"baseDoc,omitempty"`
}

// toResponse は domain.PageSuggestion を応答形へ変換する。scope.workspaceID を BaseDoc の
// GetPageVersionUseCase 呼び出しに使う（提案自体は WorkspaceID を持つが、呼び出し元が
// 既に検証済みの scope をそのまま使う方が、handler 内の他の変換と作法が揃う）。
func (h *PageSuggestionHandler) toResponse(
	ctx context.Context, scope kbRequestScope, s domain.PageSuggestion, cache pageSuggestionNameCache,
) kbPageSuggestionResponse {
	resp := kbPageSuggestionResponse{
		ID:         s.ID,
		BaseSeq:    s.BaseSeq,
		Doc:        json.RawMessage(s.Doc),
		Status:     string(s.Status),
		Author:     h.resolveRef(ctx, s.AuthorUserID, cache),
		CreatedAt:  s.CreatedAt,
		ResolvedAt: s.ResolvedAt,
	}
	if s.ResolvedByUserID != nil {
		ref := h.resolveRef(ctx, *s.ResolvedByUserID, cache)
		resp.ResolvedBy = &ref
	}
	if s.BaseSeq != nil {
		v, err := h.getVersion.Execute(ctx, kb.GetPageVersionInput{
			WorkspaceID: scope.workspaceID, PageID: s.PageID, Seq: *s.BaseSeq,
		})
		if err != nil {
			slog.WarnContext(ctx, "page suggestion: base version lookup failed", "err", err)
		} else {
			resp.BaseDoc = json.RawMessage(v.Doc)
		}
	}
	return resp
}

// kbCreateSuggestionRequest は提案作成の入力。doc の中身の検証は
// kb.CreateSuggestionUseCase（ReplacePageBlocksUseCase と同じ検証パイプライン）が行う。
type kbCreateSuggestionRequest struct {
	Doc json.RawMessage `json:"doc" binding:"required"`
}

// Create は commenter が保存した本文を提案として積む（CanComment が要る）。
func (h *PageSuggestionHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requireCommentPermissionWith(c, h.check, scope, pageID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	out, err := h.create.Execute(c.Request.Context(), kb.CreateSuggestionInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		Doc:          string(req.Doc),
		AuthorUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := pageSuggestionNameCache{}
	c.JSON(http.StatusCreated, h.toResponse(c.Request.Context(), scope, *out, cache))
}

// ListOpen はページの open な提案一覧を返す（CanView だけで許可する — コメント一覧が
// 誰でも見られるのと同じ考え方。書ける・採用できるのは commenter / editor 以上、という区別）。
func (h *PageSuggestionHandler) ListOpen(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityView) {
		return
	}
	out, err := h.listOpen.Execute(c.Request.Context(), kb.ListOpenPageSuggestionsInput{
		WorkspaceID: scope.workspaceID, PageID: pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := pageSuggestionNameCache{}
	// 0 件でも [] を返す（PageTemplateHandler.List と同じ理由 — null だとフロントの .map が落ちる）。
	items := make([]kbPageSuggestionResponse, 0, len(out))
	for _, s := range out {
		items = append(items, h.toResponse(c.Request.Context(), scope, s, cache))
	}
	c.JSON(http.StatusOK, items)
}

// Accept は提案を採用する（CanEdit が要る） — 本文へ反映し版を 1 つ切る。
func (h *PageSuggestionHandler) Accept(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityEdit) {
		return
	}
	suggestionID := c.Param("suggestionId")
	out, err := h.accept.Execute(c.Request.Context(), kb.AcceptSuggestionInput{
		WorkspaceID:    scope.workspaceID,
		PageID:         pageID,
		SuggestionID:   suggestionID,
		ResolverUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := pageSuggestionNameCache{}
	c.JSON(http.StatusOK, h.toResponse(c.Request.Context(), scope, *out, cache))
}

// Reject は提案を却下する（CanEdit が要る） — 本文は一切変えない。
func (h *PageSuggestionHandler) Reject(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityEdit) {
		return
	}
	suggestionID := c.Param("suggestionId")
	out, err := h.reject.Execute(c.Request.Context(), kb.RejectSuggestionInput{
		WorkspaceID:    scope.workspaceID,
		PageID:         pageID,
		SuggestionID:   suggestionID,
		ResolverUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := pageSuggestionNameCache{}
	c.JSON(http.StatusOK, h.toResponse(c.Request.Context(), scope, *out, cache))
}
