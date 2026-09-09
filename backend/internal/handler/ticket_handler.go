package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
)

// TicketHandler はチケット本体の操作を受ける（有効化・作成・取得・一覧・更新・並び替え・
// アーカイブ・状態変更・親変更・担当・履歴）。状態/種別マスタの管理は TicketStatusHandler /
// TicketTypeHandler が別に持つ（設計の表が違う。1 handler 1 概念）。
//
// チケットの実効権限はページを介さない「スペース単位」の判定（設計 Ⅳ-H）。対象がまだ
// 存在しない操作（一覧・作成・有効化）は checkSpace（kb パッケージの
// CheckSpacePermissionUseCase をそのまま流用。ページを一切見ないので usecase/ticket から
// usecase/kb を import しなくても handler 層でなら両方使える）、チケットを名指しする操作は
// checkTicket（ticket.CheckTicketPermissionUseCase。内部で FindTicket → スペース解決する）
// で判定する。
type TicketHandler struct {
	checkSpace    *kb.CheckSpacePermissionUseCase
	checkTicket   *ticket.CheckTicketPermissionUseCase
	resolveKey    *ticket.ResolveTicketKeyUseCase
	enable        *ticket.EnableTicketsForSpaceUseCase
	create        *ticket.CreateTicketUseCase
	get           *ticket.GetTicketUseCase
	getAssignment *ticket.GetTicketAssignmentUseCase
	list          *ticket.ListTicketsUseCase
	update        *ticket.UpdateTicketUseCase
	move          *ticket.MoveTicketUseCase
	archive       *ticket.ArchiveTicketUseCase
	restore       *ticket.RestoreTicketUseCase
	changeStat    *ticket.ChangeTicketStatusUseCase
	changeParen   *ticket.ChangeTicketParentUseCase
	assign        *ticket.AssignTicketUseCase
	unassign      *ticket.UnassignTicketUseCase
	history       *ticket.ListTicketHistoryUseCase
}

func NewTicketHandler(
	checkSpace *kb.CheckSpacePermissionUseCase,
	checkTicket *ticket.CheckTicketPermissionUseCase,
	resolveKey *ticket.ResolveTicketKeyUseCase,
	enable *ticket.EnableTicketsForSpaceUseCase,
	create *ticket.CreateTicketUseCase,
	get *ticket.GetTicketUseCase,
	getAssignment *ticket.GetTicketAssignmentUseCase,
	list *ticket.ListTicketsUseCase,
	update *ticket.UpdateTicketUseCase,
	move *ticket.MoveTicketUseCase,
	archive *ticket.ArchiveTicketUseCase,
	restore *ticket.RestoreTicketUseCase,
	changeStat *ticket.ChangeTicketStatusUseCase,
	changeParent *ticket.ChangeTicketParentUseCase,
	assign *ticket.AssignTicketUseCase,
	unassign *ticket.UnassignTicketUseCase,
	history *ticket.ListTicketHistoryUseCase,
) *TicketHandler {
	return &TicketHandler{
		checkSpace: checkSpace, checkTicket: checkTicket, resolveKey: resolveKey,
		enable: enable, create: create, get: get, getAssignment: getAssignment,
		list: list, update: update,
		move: move, archive: archive, restore: restore, changeStat: changeStat,
		changeParen: changeParent, assign: assign, unassign: unassign, history: history,
	}
}

// maxTicketBodyBytes はチケット API のボディ上限。本文は ProseMirror の JSON なので
// ナレッジページ本文 API（maxKnowledgeBaseBodyBytes）と同じ桁で足りる。
const maxTicketBodyBytes = maxKnowledgeBaseBodyBytes

// ticketEmptyDoc は本文省略時の既定値（空の ProseMirror doc）。
const ticketEmptyDoc = `{"type":"doc","content":[]}`

func limitTicketBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTicketBodyBytes)
}

// respondTicketErr は usecase / repository / domain のセンチネルを HTTP ステータスへ対応づける。
// 「存在しない」と「見る権限が無い」を同じ 404 に揃える方針は kb と同じ（respondKnowledgeBaseErr
// 参照）。チケットの実効権限はスペース単位で、ページのような個票の grant を持たないので、
// 撃ち分けの検討事項自体が kb より少ない。
func respondTicketErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrTicketNotFound),
		errors.Is(err, repository.ErrTicketStatusNotFound),
		errors.Is(err, repository.ErrTicketTypeNotFound),
		errors.Is(err, repository.ErrSpaceNotFound),
		errors.Is(err, repository.ErrWorkspaceNotFound):
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
	case errors.Is(err, repository.ErrTicketsAlreadyEnabled):
		c.JSON(http.StatusConflict, errorResponse{Error: "tickets_already_enabled"})
	case errors.Is(err, repository.ErrTicketStatusNameTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "status_name_taken"})
	case errors.Is(err, repository.ErrTicketTypeNameTaken):
		c.JSON(http.StatusConflict, errorResponse{Error: "type_name_taken"})
	case errors.Is(err, ticket.ErrTicketStatusInUse):
		c.JSON(http.StatusConflict, errorResponse{Error: "status_in_use"})
	case errors.Is(err, ticket.ErrTicketTypeInUse):
		c.JSON(http.StatusConflict, errorResponse{Error: "type_in_use"})
	case errors.Is(err, repository.ErrTicketAssigneeNotFound):
		// 担当に指定した principal がこのワークスペースに実在しない（別ワークスペース /
		// kind != user を含む）。リクエスト本文の値が悪いので、URL の対象を隠す 404 群とは
		// 分け、400 として返す（担当候補の実在は ListGrantablePrincipals で既に見えており、
		// 隠す意味が無い）。
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_assignee"})
	case errors.Is(err, domain.ErrTicketHierarchyRejected):
		c.JSON(http.StatusConflict, errorResponse{Error: "ticket_hierarchy_rejected"})
	case errors.Is(err, domain.ErrTicketDateRangeInverted):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_date_range"})
	case errors.Is(err, ticket.ErrTicketMoveAnchorNotSibling):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "anchor_not_sibling"})
	case errors.Is(err, domain.ErrInvalidTicketName),
		errors.Is(err, domain.ErrInvalidTicketColor),
		errors.Is(err, domain.ErrInvalidTicketStatusCategory),
		errors.Is(err, domain.ErrInvalidTicketHierarchyLevel):
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
	default:
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
	}
}

// requireTicketSpacePermission はスペース単位の実効権限を確かめる（対象がまだ存在しない
// 操作専用。requireTicketPermission と同じ 404/403 の撃ち分けをする）。
func (h *TicketHandler) requireTicketSpacePermission(
	c *gin.Context, scope kbRequestScope, spaceID string, capability domain.Capability,
) bool {
	return requireTicketSpacePermissionWith(c, h.checkSpace, scope, spaceID, capability)
}

// requireTicketSpacePermissionWith は requireTicketSpacePermission の実体。TicketHandler /
// TicketStatusHandler / TicketTypeHandler の 3 つが同じ判定（スペース単位・対象がまだ
// 存在しない操作の入口）を使うために package レベルの関数へ切り出してある
// （kb の requirePagePermissionWith と同じ理由 — 書き直すとどれか 1 つだけ直し忘れて食い違う）。
func requireTicketSpacePermissionWith(
	c *gin.Context, checkSpace *kb.CheckSpacePermissionUseCase,
	scope kbRequestScope, spaceID string, capability domain.Capability,
) bool {
	perm, err := checkSpace.Execute(c.Request.Context(), kb.CheckSpacePermissionInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID, UserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return false
	}
	return requireScopeCapability(c, perm, capability)
}

// requireTicketPermission はチケット 1 件の実効権限を確かめる（スペース単位の判定を
// CheckTicketPermissionUseCase 経由で行う。ページ付与のような個票の例外は無い）。
func (h *TicketHandler) requireTicketPermission(
	c *gin.Context, scope kbRequestScope, ticketID string, capability domain.Capability,
) bool {
	perm, err := h.checkTicket.Execute(c.Request.Context(), ticket.CheckTicketPermissionInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, UserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return false
	}
	return requireScopeCapability(c, perm, capability)
}

// requireScopeCapability は ScopePermission から 404/403 を書き分ける共通の末尾処理。
// TicketHandler の 2 つの入口（スペース単位・チケット単位）が同じ規則を使うために
// package レベルではなく型に閉じたヘルパーへ切り出してある。
func requireScopeCapability(c *gin.Context, perm *domain.ScopePermission, capability domain.Capability) bool {
	if !perm.CanView {
		c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
		return false
	}
	if !perm.Allows(capability) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// ticketEnableRequest は有効化の入力。SourceSpaceID を指定すると、そのスペースの現役構成を
// 複製する（設計 Ⅵ）。
type ticketEnableRequest struct {
	SourceSpaceID string `json:"sourceSpaceId,omitempty"`
}

// Enable はスペースにチケット機能を有効化する（スペースの編集権限が要る）。
func (h *TicketHandler) Enable(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireTicketSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
		return
	}
	// ボディは省略できる（最小構成で有効化する既定の経路）。ShouldBindJSON は空ボディを
	// io.EOF にするので、それだけは無視して既定値（SourceSpaceID なし）のまま進む。
	// 壊れた JSON（EOF ではない）はふつうに 400 で断る。
	var req ticketEnableRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
			return
		}
	}
	var sourceSpaceID *string
	if req.SourceSpaceID != "" {
		// 複製元スペースは自分が閲覧できるものに限る（他社テナントのスペース構成を
		// 覗き見る経路にしない）。
		if !h.requireTicketSpacePermission(c, scope, req.SourceSpaceID, domain.CapabilityView) {
			return
		}
		sourceSpaceID = &req.SourceSpaceID
	}
	out, err := h.enable.Execute(c.Request.Context(), ticket.EnableTicketsForSpaceInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID, SourceSpaceID: sourceSpaceID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, out)
}

// ticketCreateRequest はチケット作成の入力。TypeID / StatusID / Doc は省略でき、
// 省略時はそれぞれ既定種別・初期状態・空の本文になる（CreateTicketUseCase 参照）。
type ticketCreateRequest struct {
	ParentID  string          `json:"parentId,omitempty"`
	TypeID    string          `json:"typeId,omitempty"`
	StatusID  string          `json:"statusId,omitempty"`
	Title     string          `json:"title" binding:"required,max=200"`
	Doc       json.RawMessage `json:"doc,omitempty"`
	Priority  int             `json:"priority,omitempty"`
	StartDate *string         `json:"startDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
	DueDate   *string         `json:"dueDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
}

// Create はスペース直下（または親チケットの下）に新しいチケットを作る（スペースの
// 編集権限が要る。親を名指しした場合でも、チケットはページのような個票の権限を
// 持たないので判定はスペース単位のまま — 親の実在・同一スペースは usecase 側で検証する）。
func (h *TicketHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireTicketSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
		return
	}
	limitTicketBody(c)
	var req ticketCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	if req.Priority != 0 && !domain.TicketPriority(req.Priority).Valid() {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	doc := req.Doc
	if len(doc) == 0 {
		doc = json.RawMessage(ticketEmptyDoc)
	}
	var parentID *string
	if req.ParentID != "" {
		parentID = &req.ParentID
	}
	t, err := h.create.Execute(c.Request.Context(), ticket.CreateTicketInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID,
		TypeID: req.TypeID, StatusID: req.StatusID, ParentID: parentID,
		Title: req.Title, Doc: string(doc), Priority: domain.TicketPriority(req.Priority),
		StartDate: req.StartDate, DueDate: req.DueDate, CreatedByUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusCreated)
}

// Get はチケット 1 件を返す（閲覧権限が要る）。
func (h *TicketHandler) Get(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityView) {
		return
	}
	found, err := h.get.Execute(c.Request.Context(), ticket.GetTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticketResponse{
		Ticket: &found.Ticket, AssigneePrincipalID: found.AssigneePrincipalID,
	})
}

// ResolveByKey は表示キー（例 FRESTYLE-12）からチケット 1 件を返す（閲覧権限が要る）。
// キーの分解に失敗した場合も実在しない場合と同じ 404 にする
// （ResolveTicketKeyUseCase が両方を repository.ErrTicketNotFound へ畳んでいる）。
func (h *TicketHandler) ResolveByKey(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID, err := h.resolveKey.Execute(c.Request.Context(), ticket.ResolveTicketKeyInput{
		WorkspaceID: scope.workspaceID, Key: c.Param("key"),
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityView) {
		return
	}
	found, err := h.get.Execute(c.Request.Context(), ticket.GetTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticketResponse{
		Ticket: &found.Ticket, AssigneePrincipalID: found.AssigneePrincipalID,
	})
}

// ticketResponse はチケット 1 件の返却形。
//
// domain.Ticket をそのまま埋め込み（JSON は平らに出る）、別表にある担当だけを足す。
// 一覧・詳細・変更系のすべてがこの 1 つの形で返るので、画面は応答の出どころで
// 型を出し分けなくてよい（担当が居なければ assigneePrincipalId は出ない）。
type ticketResponse struct {
	*domain.Ticket
	AssigneePrincipalID *string `json:"assigneePrincipalId,omitempty"`
}

// ticketListResponse は一覧の返却形。
type ticketListResponse struct {
	Tickets []ticketResponse `json:"tickets"`
}

// respondTicket は変更系の応答を組み立てて返す。担当は usecase が触らないので、
// ここで 1 回だけ引いて詰める（引けなければ担当なしとして返し、応答自体は止めない —
// 変更そのものは既に成功しているため。kb が最終編集者の名前で採るのと同じ扱い）。
func (h *TicketHandler) respondTicket(c *gin.Context, scope kbRequestScope, t *domain.Ticket, status int) {
	res := ticketResponse{Ticket: t}
	a, err := h.getAssignment.Execute(c.Request.Context(), scope.workspaceID, t.ID)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "ticket: assignee lookup failed", "err", err, "ticketId", t.ID)
	} else if a != nil {
		id := a.AssigneePrincipalID
		res.AssigneePrincipalID = &id
	}
	c.JSON(status, res)
}

// List はスペース内のチケット一覧を返す（スペースの閲覧権限が要る）。
func (h *TicketHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireTicketSpacePermission(c, scope, spaceID, domain.CapabilityView) {
		return
	}
	var statusID, typeID, assigneeID *string
	if v := c.Query("statusId"); v != "" {
		statusID = &v
	}
	if v := c.Query("typeId"); v != "" {
		typeID = &v
	}
	if v := c.Query("assigneePrincipalId"); v != "" {
		assigneeID = &v
	}
	tickets, err := h.list.Execute(c.Request.Context(), ticket.ListTicketsInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID,
		IncludeArchived: c.Query("archived") == "true",
		StatusID:        statusID, TypeID: typeID, AssigneePrincipalID: assigneeID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	out := make([]ticketResponse, 0, len(tickets))
	for i := range tickets {
		out = append(out, ticketResponse{
			Ticket:              &tickets[i].Ticket,
			AssigneePrincipalID: tickets[i].AssigneePrincipalID,
		})
	}
	c.JSON(http.StatusOK, ticketListResponse{Tickets: out})
}

// ticketUpdateRequest はチケット更新の入力（PUT 相当。呼び出し側は現在の望ましい値を
// 毎回すべて渡す — UpdateTicketUseCase の doc 参照）。
type ticketUpdateRequest struct {
	Title     string          `json:"title" binding:"required,max=200"`
	Doc       json.RawMessage `json:"doc" binding:"required"`
	TypeID    string          `json:"typeId" binding:"required"`
	Priority  int             `json:"priority" binding:"required,oneof=1 2 3"`
	StartDate *string         `json:"startDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
	DueDate   *string         `json:"dueDate,omitempty" binding:"omitempty,datetime=2006-01-02"`
}

// Update はチケットの title / doc / type / priority / 日付を書き換える（編集権限が要る）。
func (h *TicketHandler) Update(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	limitTicketBody(c)
	var req ticketUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	t, err := h.update.Execute(c.Request.Context(), ticket.UpdateTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
		Title: req.Title, Doc: string(req.Doc), TypeID: req.TypeID,
		Priority:  domain.TicketPriority(req.Priority),
		StartDate: req.StartDate, DueDate: req.DueDate, ActorUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusOK)
}

// ticketMoveRequest は並び替えの入力。AnchorTicketID を省略すると末尾に置く。
type ticketMoveRequest struct {
	AnchorTicketID string `json:"anchorTicketId,omitempty"`
	AnchorAfter    bool   `json:"anchorAfter,omitempty"`
}

// Move はチケットの並び順を変える（編集権限が要る）。
func (h *TicketHandler) Move(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	limitTicketBody(c)
	var req ticketMoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	var anchor *string
	if req.AnchorTicketID != "" {
		// 隣に指定したチケットは閲覧できなければならない（kb の Move と同じ理由 —
		// 編集できれば誰でも叩ける口で、実在を無条件に言い当てさせない）。
		if !h.requireTicketPermission(c, scope, req.AnchorTicketID, domain.CapabilityView) {
			return
		}
		anchor = &req.AnchorTicketID
	}
	if err := h.move.Execute(c.Request.Context(), ticket.MoveTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
		AnchorTicketID: anchor, AnchorAfter: req.AnchorAfter,
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Archive はチケットをアーカイブする（編集権限が要る）。
func (h *TicketHandler) Archive(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	t, err := h.archive.Execute(c.Request.Context(), ticket.ArchiveTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, ActorUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusOK)
}

// Restore はアーカイブ済みチケットを現役へ戻す（編集権限が要る）。
func (h *TicketHandler) Restore(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	t, err := h.restore.Execute(c.Request.Context(), ticket.RestoreTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, ActorUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusOK)
}

// ticketChangeStatusRequest は状態変更の入力。Resolution は category=done のときだけ使う
// （ChangeTicketStatusUseCase / domain.ResolveTicketClosedFields が導出する）。
type ticketChangeStatusRequest struct {
	StatusID   string  `json:"statusId" binding:"required"`
	Resolution *string `json:"resolution,omitempty"`
}

// ChangeStatus はチケットの状態を変える（編集権限が要る）。
func (h *TicketHandler) ChangeStatus(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	var req ticketChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	var resolution *domain.TicketResolution
	if req.Resolution != nil {
		r := domain.TicketResolution(*req.Resolution)
		if !r.Valid() {
			c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
			return
		}
		resolution = &r
	}
	t, err := h.changeStat.Execute(c.Request.Context(), ticket.ChangeTicketStatusInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, StatusID: req.StatusID,
		Resolution: resolution, ActorUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusOK)
}

// ticketChangeParentRequest は親変更の入力。ParentID を省略するとトップレベルへ戻す。
type ticketChangeParentRequest struct {
	ParentID string `json:"parentId,omitempty"`
}

// ChangeParent はチケットの親を変える（編集権限が要る）。
func (h *TicketHandler) ChangeParent(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	var req ticketChangeParentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	var newParentID *string
	if req.ParentID != "" {
		// 新しい親は編集できなければならない（kb の Move と同じ理由。書けないサブツリーへ
		// 差し込めてしまうのを防ぐ）。
		if !h.requireTicketPermission(c, scope, req.ParentID, domain.CapabilityEdit) {
			return
		}
		newParentID = &req.ParentID
	}
	t, err := h.changeParen.Execute(c.Request.Context(), ticket.ChangeTicketParentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
		NewParentID: newParentID, ActorUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	h.respondTicket(c, scope, t, http.StatusOK)
}

// ticketAssignRequest は担当設定の入力。
type ticketAssignRequest struct {
	AssigneePrincipalID string `json:"assigneePrincipalId" binding:"required"`
}

// Assign はチケットの担当者を設定する（編集権限が要る）。
func (h *TicketHandler) Assign(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	var req ticketAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	a, err := h.assign.Execute(c.Request.Context(), ticket.AssignTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
		AssigneePrincipalID: req.AssigneePrincipalID, AssignedByUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, a)
}

// Unassign はチケットの担当を外す（編集権限が要る）。
func (h *TicketHandler) Unassign(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	if err := h.unassign.Execute(c.Request.Context(), ticket.UnassignTicketInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, ActorUserID: scope.userID,
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ticketHistoryResponse は変更履歴の返却形。
type ticketHistoryResponse struct {
	Groups []domain.TicketChangeGroup `json:"groups"`
}

// History はチケットの変更履歴を新しい順に返す（閲覧権限が要る）。
func (h *TicketHandler) History(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityView) {
		return
	}
	groups, err := h.history.Execute(c.Request.Context(), ticket.ListTicketHistoryInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	if groups == nil {
		groups = []domain.TicketChangeGroup{}
	}
	c.JSON(http.StatusOK, ticketHistoryResponse{Groups: groups})
}
