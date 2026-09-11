package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
)

// TicketCommentHandler はチケットへの発言・編集履歴・反応を受ける（段 3）。
// 状態遷移ログ（ticket_status_transitions）は ChangeTicketStatusUseCase が書くだけで
// 専用の API を持たない（着手時に解く未決: 集計 API は後段へ回す判断）。
type TicketCommentHandler struct {
	checkTicket    *ticket.CheckTicketPermissionUseCase
	create         *ticket.CreateTicketCommentUseCase
	update         *ticket.UpdateTicketCommentUseCase
	del            *ticket.DeleteTicketCommentUseCase
	list           *ticket.ListTicketCommentsUseCase
	listEdits      *ticket.ListTicketCommentEditsUseCase
	addReaction    *ticket.AddTicketCommentReactionUseCase
	removeReaction *ticket.RemoveTicketCommentReactionUseCase
	userName       *kb.LookupUserNameUseCase
}

func NewTicketCommentHandler(
	checkTicket *ticket.CheckTicketPermissionUseCase,
	create *ticket.CreateTicketCommentUseCase,
	update *ticket.UpdateTicketCommentUseCase,
	del *ticket.DeleteTicketCommentUseCase,
	list *ticket.ListTicketCommentsUseCase,
	listEdits *ticket.ListTicketCommentEditsUseCase,
	addReaction *ticket.AddTicketCommentReactionUseCase,
	removeReaction *ticket.RemoveTicketCommentReactionUseCase,
	userName *kb.LookupUserNameUseCase,
) *TicketCommentHandler {
	return &TicketCommentHandler{
		checkTicket: checkTicket, create: create, update: update, del: del,
		list: list, listEdits: listEdits,
		addReaction: addReaction, removeReaction: removeReaction, userName: userName,
	}
}

// requireTicketCommentScope はチケット単位の実効権限を引く。CanComment は Capability
// （View/Edit の 2 値）に無い別軸のフィールドなので、呼び出し側が用途に応じて
// perm.CanComment / perm.CanManage を直接見る（ticket_handler.go の
// requireTicketPermission は Capability しか見られず、ここでは使えない）。
func (h *TicketCommentHandler) requireTicketCommentScope(c *gin.Context, scope kbRequestScope, ticketID string) (*domain.ScopePermission, bool) {
	perm, err := h.checkTicket.Execute(c.Request.Context(), ticket.CheckTicketPermissionInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, UserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return nil, false
	}
	if !perm.CanView {
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return nil, false
	}
	return perm, true
}

type ticketCommentAuthorResponse struct {
	UserID uint64 `json:"userId"`
	Name   string `json:"name"`
}

type ticketCommentReactionResponse struct {
	UserID uint64 `json:"userId"`
	Emoji  string `json:"emoji"`
}

type ticketCommentResponse struct {
	ID              string                          `json:"id"`
	ParentCommentID *string                         `json:"parentCommentId,omitempty"`
	Author          ticketCommentAuthorResponse     `json:"author"`
	Body            json.RawMessage                 `json:"body"`
	Edited          bool                            `json:"edited"`
	Reactions       []ticketCommentReactionResponse `json:"reactions"`
	CreatedAt       time.Time                       `json:"createdAt"`
	UpdatedAt       time.Time                       `json:"updatedAt"`
}

type ticketCommentListResponse struct {
	Comments []ticketCommentResponse `json:"comments"`
}

type ticketCommentNameCache map[uint64]string

func (h *TicketCommentHandler) resolveAuthor(c *gin.Context, userID uint64, cache ticketCommentNameCache) ticketCommentAuthorResponse {
	name, ok := cache[userID]
	if !ok {
		var err error
		name, err = h.userName.Execute(c.Request.Context(), userID)
		if err != nil {
			slog.WarnContext(c.Request.Context(), "ticket comment: author name resolve failed", "err", err)
		}
		cache[userID] = name
	}
	return ticketCommentAuthorResponse{UserID: userID, Name: name}
}

func (h *TicketCommentHandler) toResponse(
	c *gin.Context, item ticket.TicketCommentWithReactions, cache ticketCommentNameCache,
) ticketCommentResponse {
	reactions := make([]ticketCommentReactionResponse, 0, len(item.Reactions))
	for _, r := range item.Reactions {
		reactions = append(reactions, ticketCommentReactionResponse{UserID: r.UserID, Emoji: r.Emoji})
	}
	return ticketCommentResponse{
		ID:              item.Comment.ID,
		ParentCommentID: item.Comment.ParentCommentID,
		Author:          h.resolveAuthor(c, item.Comment.AuthorUserID, cache),
		Body:            json.RawMessage(item.Comment.Body),
		Edited:          item.Comment.EditedAt != nil,
		Reactions:       reactions,
		CreatedAt:       item.Comment.CreatedAt,
		UpdatedAt:       item.Comment.UpdatedAt,
	}
}

// List はチケットの発言一覧を返す（閲覧できれば読める。返信は parentCommentId を見て
// 画面がツリーへ組み立てる）。
func (h *TicketCommentHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if _, ok := h.requireTicketCommentScope(c, scope, ticketID); !ok {
		return
	}
	out, err := h.list.Execute(c.Request.Context(), scope.workspaceID, ticketID)
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	cache := ticketCommentNameCache{}
	comments := make([]ticketCommentResponse, 0, len(out))
	for _, item := range out {
		comments = append(comments, h.toResponse(c, item, cache))
	}
	c.JSON(http.StatusOK, ticketCommentListResponse{Comments: comments})
}

type ticketCommentCreateRequest struct {
	ParentCommentID string          `json:"parentCommentId,omitempty"`
	Body            json.RawMessage `json:"body" binding:"required"`
}

// Create は発言（返信を含む）を 1 件作る（CanComment が要る）。
func (h *TicketCommentHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	perm, ok := h.requireTicketCommentScope(c, scope, ticketID)
	if !ok {
		return
	}
	if !perm.CanComment {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	limitTicketBody(c)
	var req ticketCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	var parentID *string
	if req.ParentCommentID != "" {
		parentID = &req.ParentCommentID
	}
	created, err := h.create.Execute(c.Request.Context(), ticket.CreateTicketCommentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, ParentCommentID: parentID,
		AuthorUserID: scope.userID, Body: string(req.Body),
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	cache := ticketCommentNameCache{}
	c.JSON(http.StatusCreated, h.toResponse(c, ticket.TicketCommentWithReactions{Comment: *created}, cache))
}

type ticketCommentUpdateRequest struct {
	Body json.RawMessage `json:"body" binding:"required"`
}

// Update は発言の本文を書き換える（投稿者本人か CanManage が要る）。
func (h *TicketCommentHandler) Update(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	perm, ok := h.requireTicketCommentScope(c, scope, ticketID)
	if !ok {
		return
	}
	commentID := c.Param("commentId")
	limitTicketBody(c)
	var req ticketCommentUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	updated, err := h.update.Execute(c.Request.Context(), ticket.UpdateTicketCommentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, CommentID: commentID,
		ActorUserID: scope.userID, ActorCanManage: perm.CanManage, Body: string(req.Body),
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	cache := ticketCommentNameCache{}
	c.JSON(http.StatusOK, h.toResponse(c, ticket.TicketCommentWithReactions{Comment: *updated}, cache))
}

// Delete は発言を消えたことにする（投稿者本人か CanManage が要る）。
func (h *TicketCommentHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	perm, ok := h.requireTicketCommentScope(c, scope, ticketID)
	if !ok {
		return
	}
	commentID := c.Param("commentId")
	if err := h.del.Execute(c.Request.Context(), ticket.DeleteTicketCommentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, CommentID: commentID,
		ActorUserID: scope.userID, ActorCanManage: perm.CanManage,
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type ticketCommentEditResponse struct {
	ID           string                      `json:"id"`
	Editor       ticketCommentAuthorResponse `json:"editor"`
	PreviousBody json.RawMessage             `json:"previousBody"`
	EditedAt     time.Time                   `json:"editedAt"`
}

type ticketCommentEditListResponse struct {
	Edits []ticketCommentEditResponse `json:"edits"`
}

// ListEdits は発言 1 件の編集履歴を返す（閲覧できれば読める）。
func (h *TicketCommentHandler) ListEdits(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if _, ok := h.requireTicketCommentScope(c, scope, ticketID); !ok {
		return
	}
	commentID := c.Param("commentId")
	out, err := h.listEdits.Execute(c.Request.Context(), scope.workspaceID, ticketID, commentID)
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	cache := ticketCommentNameCache{}
	edits := make([]ticketCommentEditResponse, 0, len(out))
	for _, e := range out {
		edits = append(edits, ticketCommentEditResponse{
			ID: e.ID, Editor: h.resolveAuthor(c, e.EditorUserID, cache),
			PreviousBody: json.RawMessage(e.PreviousBody), EditedAt: e.EditedAt,
		})
	}
	c.JSON(http.StatusOK, ticketCommentEditListResponse{Edits: edits})
}

// AddReaction は発言へ絵文字反応を付ける（CanComment が要る。付け外しは冪等）。
func (h *TicketCommentHandler) AddReaction(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	perm, ok := h.requireTicketCommentScope(c, scope, ticketID)
	if !ok {
		return
	}
	if !perm.CanComment {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	commentID, emoji := c.Param("commentId"), c.Param("emoji")
	if err := h.addReaction.Execute(c.Request.Context(), scope.workspaceID, ticketID, commentID, scope.userID, emoji); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveReaction は自分が付けた反応を外す（CanComment が要る）。
func (h *TicketCommentHandler) RemoveReaction(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	perm, ok := h.requireTicketCommentScope(c, scope, ticketID)
	if !ok {
		return
	}
	if !perm.CanComment {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	commentID, emoji := c.Param("commentId"), c.Param("emoji")
	if err := h.removeReaction.Execute(c.Request.Context(), scope.workspaceID, ticketID, commentID, scope.userID, emoji); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
