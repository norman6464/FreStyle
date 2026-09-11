package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/comment"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
	"github.com/norman6464/frestyle/backend/internal/usecase/user"
)

// CommentHandler はページ全体へのコメントと錨付きコメントを受ける。
// 錨付け（block_id / anchor_from / anchor_to / quote）を受け取るのは CreateThread だけ
// （返信・解決・再開はスレッド単位の操作で錨を持たない）。
type CommentHandler struct {
	check        *kb.CheckPagePermissionUseCase
	createThread *comment.CreateCommentThreadUseCase
	addComment   *comment.AddCommentUseCase
	listThreads  *comment.ListCommentThreadsUseCase
	resolve      *comment.ResolveCommentThreadUseCase
	reopen       *comment.ReopenCommentThreadUseCase
	userDisplay  *user.LookupUserDisplayUseCase
}

// NewCommentHandler は CommentHandler を組み立てる。
func NewCommentHandler(
	check *kb.CheckPagePermissionUseCase,
	createThread *comment.CreateCommentThreadUseCase,
	addComment *comment.AddCommentUseCase,
	listThreads *comment.ListCommentThreadsUseCase,
	resolve *comment.ResolveCommentThreadUseCase,
	reopen *comment.ReopenCommentThreadUseCase,
	userDisplay *user.LookupUserDisplayUseCase,
) *CommentHandler {
	return &CommentHandler{
		check:        check,
		createThread: createThread,
		addComment:   addComment,
		listThreads:  listThreads,
		resolve:      resolve,
		reopen:       reopen,
		userDisplay:  userDisplay,
	}
}

// requireCommentPermission はコメント操作の実効権限を確かめる（CanComment）。
// 満たさなければレスポンスを書いて false を返す。
//
// requirePagePermission（kb_page_handler.go）と同じ形だが、判定するのが
// domain.Capability ではなく domain.PagePermission.CanComment という別軸のフィールドなので、
// 別関数にする（Capability は view/edit の 2 値にしか対応しておらず、そのまま流用できない）。
func (h *CommentHandler) requireCommentPermission(c *gin.Context, scope kbRequestScope, pageID string) bool {
	return requireCommentPermissionWith(c, h.check, scope, pageID)
}

// requireCommentPermissionWith は requireCommentPermission の実体。CommentHandler と
// PageSuggestionHandler の両方が同じ判定（CanComment）を使うために package レベルの関数へ
// 切り出してある（requirePagePermissionWith / requireSpacePermissionWith と同じ理由 —
// CanComment の判定はどちらの handler でも同じで、書き直すとどちらか片方だけ直し忘れて
// 食い違う危険がある）。
func requireCommentPermissionWith(c *gin.Context, check *kb.CheckPagePermissionUseCase, scope kbRequestScope, pageID string) bool {
	perm, err := check.Execute(c.Request.Context(), kb.CheckPagePermissionInput{
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
	if !perm.CanComment {
		// ここに来る相手は閲覧できる = 実在を既に知っているので、403 で理由を返してよい。
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// commentResponse は発言 1 件の返却形（最初の発言も返信も同じ形）。
type commentResponse struct {
	ID        string              `json:"id"`
	Author    userDisplayResponse `json:"author"`
	Body      json.RawMessage     `json:"body"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

// commentThreadResponse はスレッド 1 件と、その発言（最初の発言 + 返信）の返却形。
//
// BlockID/AnchorFrom/AnchorTo/Quote は錨付きスレッドだけ値を持つ。page-level の
// スレッドでは 4 つとも省略される（omitempty）。
type commentThreadResponse struct {
	ID         string               `json:"id"`
	CreatedBy  userDisplayResponse  `json:"createdBy"`
	BlockID    *string              `json:"blockId,omitempty"`
	AnchorFrom *int                 `json:"anchorFrom,omitempty"`
	AnchorTo   *int                 `json:"anchorTo,omitempty"`
	Quote      *string              `json:"quote,omitempty"`
	ResolvedAt *time.Time           `json:"resolvedAt,omitempty"`
	ResolvedBy *userDisplayResponse `json:"resolvedBy,omitempty"`
	CreatedAt  time.Time            `json:"createdAt"`
	UpdatedAt  time.Time            `json:"updatedAt"`
	Comments   []commentResponse    `json:"comments"`
}

// toCommentResponse は domain.Comment を応答形へ変換する。
func (h *CommentHandler) toCommentResponse(ctx context.Context, c domain.Comment, cache userDisplayCache) commentResponse {
	return commentResponse{
		ID:        c.ID,
		Author:    resolveUserDisplay(ctx, h.userDisplay, c.AuthorUserID, cache),
		Body:      json.RawMessage(c.Body),
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// toCommentThreadResponse は domain.CommentThread とその発言一覧を応答形へ変換する。
func (h *CommentHandler) toCommentThreadResponse(
	ctx context.Context, t domain.CommentThread, comments []domain.Comment, cache userDisplayCache,
) commentThreadResponse {
	resp := commentThreadResponse{
		ID:         t.ID,
		CreatedBy:  resolveUserDisplay(ctx, h.userDisplay, t.CreatedByUserID, cache),
		BlockID:    t.BlockID,
		AnchorFrom: t.AnchorFrom,
		AnchorTo:   t.AnchorTo,
		Quote:      t.Quote,
		ResolvedAt: t.ResolvedAt,
		CreatedAt:  t.CreatedAt,
		UpdatedAt:  t.UpdatedAt,
		Comments:   make([]commentResponse, 0, len(comments)),
	}
	if t.ResolvedByUserID != nil {
		ref := resolveUserDisplay(ctx, h.userDisplay, *t.ResolvedByUserID, cache)
		resp.ResolvedBy = &ref
	}
	for _, c := range comments {
		resp.Comments = append(resp.Comments, h.toCommentResponse(ctx, c, cache))
	}
	return resp
}

// kbCommentBodyRequest は発言（返信）の入力。
type kbCommentBodyRequest struct {
	// Body は ProseMirror インラインノードの配列（JSON）。中身の検証は
	// domain.ValidateCommentBody（usecase 経由）が行う。
	Body json.RawMessage `json:"body" binding:"required"`
}

// kbCreateThreadRequest は新しいスレッドの入力。本文に加え、錨を任意で受け取る。
// BlockID/AnchorFrom/AnchorTo/Quote は 4 つとも揃うか 4 つとも無いかのどちらかで、
// その検証は domain.ValidateCommentAnchor（usecase 経由）が行う。
type kbCreateThreadRequest struct {
	Body       json.RawMessage `json:"body" binding:"required"`
	BlockID    *string         `json:"blockId,omitempty"`
	AnchorFrom *int            `json:"anchorFrom,omitempty"`
	AnchorTo   *int            `json:"anchorTo,omitempty"`
	Quote      *string         `json:"quote,omitempty"`
}

// CreateThread はページに新しいコメントスレッドを立てる（CanComment が要る）。
func (h *CommentHandler) CreateThread(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requireCommentPermission(c, scope, pageID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	out, err := h.createThread.Execute(c.Request.Context(), comment.CreateCommentThreadInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		AuthorUserID: scope.userID,
		Body:         string(req.Body),
		Anchor: repository.CommentAnchor{
			BlockID:    req.BlockID,
			AnchorFrom: req.AnchorFrom,
			AnchorTo:   req.AnchorTo,
			Quote:      req.Quote,
		},
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusCreated, h.toCommentThreadResponse(c.Request.Context(), out.Thread, []domain.Comment{out.Comment}, cache))
}

// AddComment は既存のスレッドへ返信を 1 件足す（CanComment が要る）。
func (h *CommentHandler) AddComment(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requireCommentPermission(c, scope, pageID) {
		return
	}
	threadID := c.Param("threadId")
	limitKnowledgeBaseBody(c)
	var req kbCommentBodyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	out, err := h.addComment.Execute(c.Request.Context(), comment.AddCommentInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		ThreadID:     threadID,
		AuthorUserID: scope.userID,
		Body:         string(req.Body),
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusCreated, h.toCommentResponse(c.Request.Context(), *out, cache))
}

// kbCommentThreadsResponse はスレッド一覧取得の応答全体。
type kbCommentThreadsResponse struct {
	Threads []commentThreadResponse `json:"threads"`
}

// ListThreads はページのコメントスレッド一覧を、それぞれの発言付きで返す
// （CanView だけで許可する — viewer でも既存のコメントは読める。書ける・解決できるのは
// commenter 以上だけ、という区別）。
func (h *CommentHandler) ListThreads(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.check, scope, pageID, domain.CapabilityView) {
		return
	}
	out, err := h.listThreads.Execute(c.Request.Context(), comment.ListCommentThreadsInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	threads := make([]commentThreadResponse, 0, len(out))
	for _, t := range out {
		threads = append(threads, h.toCommentThreadResponse(c.Request.Context(), t.Thread, t.Comments, cache))
	}
	c.JSON(http.StatusOK, kbCommentThreadsResponse{Threads: threads})
}

// Resolve はスレッドを解決済みにする（CanComment が要る）。
func (h *CommentHandler) Resolve(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requireCommentPermission(c, scope, pageID) {
		return
	}
	threadID := c.Param("threadId")
	t, err := h.resolve.Execute(c.Request.Context(), comment.ResolveCommentThreadInput{
		WorkspaceID:      scope.workspaceID,
		PageID:           pageID,
		ThreadID:         threadID,
		ResolvedByUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusOK, h.toCommentThreadResponse(c.Request.Context(), *t, nil, cache))
}

// Reopen は解決済みのスレッドを未解決へ戻す（CanComment が要る）。
func (h *CommentHandler) Reopen(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requireCommentPermission(c, scope, pageID) {
		return
	}
	threadID := c.Param("threadId")
	t, err := h.reopen.Execute(c.Request.Context(), comment.ReopenCommentThreadInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		ThreadID:    threadID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	cache := userDisplayCache{}
	c.JSON(http.StatusOK, h.toCommentThreadResponse(c.Request.Context(), *t, nil, cache))
}
