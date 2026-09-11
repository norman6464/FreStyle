package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/ticket"
)

// TicketLabelHandler はラベルの管理（スペース単位）とチケットへの付け外し（チケット単位）を
// 受ける（段 4）。管理はスペース単位の判定（TicketStatusHandler と同じ形）、付け外しは
// チケット単位の判定（TicketHandler.requireTicketPermission と同じ形）を両方使うので、
// checkSpace / checkTicket を両方持つ（TicketHandler 自身と同じ構え）。
type TicketLabelHandler struct {
	checkSpace  *kb.CheckSpacePermissionUseCase
	checkTicket *ticket.CheckTicketPermissionUseCase
	list        *ticket.ListLabelsUseCase
	create      *ticket.CreateLabelUseCase
	update      *ticket.UpdateLabelUseCase
	del         *ticket.DeleteLabelUseCase
	addToTicket *ticket.AddTicketLabelUseCase
	removeFrom  *ticket.RemoveTicketLabelUseCase
}

func NewTicketLabelHandler(
	checkSpace *kb.CheckSpacePermissionUseCase,
	checkTicket *ticket.CheckTicketPermissionUseCase,
	list *ticket.ListLabelsUseCase,
	create *ticket.CreateLabelUseCase,
	update *ticket.UpdateLabelUseCase,
	del *ticket.DeleteLabelUseCase,
	addToTicket *ticket.AddTicketLabelUseCase,
	removeFrom *ticket.RemoveTicketLabelUseCase,
) *TicketLabelHandler {
	return &TicketLabelHandler{
		checkSpace: checkSpace, checkTicket: checkTicket, list: list, create: create, update: update,
		del: del, addToTicket: addToTicket, removeFrom: removeFrom,
	}
}

func (h *TicketLabelHandler) requireSpacePermission(
	c *gin.Context, scope kbRequestScope, spaceID string, capability domain.Capability,
) bool {
	return requireTicketSpacePermissionWith(c, h.checkSpace, scope, spaceID, capability)
}

func (h *TicketLabelHandler) requireTicketPermission(
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

// ticketLabelListResponse はラベル一覧の返却形。
type ticketLabelListResponse struct {
	Labels []domain.Label `json:"labels"`
}

// List はスペースのラベル一覧を返す（閲覧権限が要る）。
func (h *TicketLabelHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpacePermission(c, scope, spaceID, domain.CapabilityView) {
		return
	}
	labels, err := h.list.Execute(c.Request.Context(), scope.workspaceID, spaceID)
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	if labels == nil {
		labels = []domain.Label{}
	}
	c.JSON(http.StatusOK, ticketLabelListResponse{Labels: labels})
}

// ticketLabelRequest はラベルの作成・更新の入力。
type ticketLabelRequest struct {
	Name  string `json:"name" binding:"required,max=64"`
	Color string `json:"color" binding:"required"`
}

// Create はスペースにラベルを 1 つ追加する（編集権限が要る）。
func (h *TicketLabelHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
		return
	}
	var req ticketLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	label, err := h.create.Execute(c.Request.Context(), ticket.CreateLabelInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID, Name: req.Name, Color: req.Color,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, label)
}

// Update はラベルの名前・色を書き換える（編集権限が要る）。
func (h *TicketLabelHandler) Update(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
		return
	}
	var req ticketLabelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	label, err := h.update.Execute(c.Request.Context(), ticket.UpdateLabelInput{
		WorkspaceID: scope.workspaceID, SpaceID: spaceID,
		LabelID: c.Param("labelId"), Name: req.Name, Color: req.Color,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, label)
}

// Delete はラベルを削除する（編集権限が要る）。
func (h *TicketLabelHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpacePermission(c, scope, spaceID, domain.CapabilityEdit) {
		return
	}
	if err := h.del.Execute(c.Request.Context(), scope.workspaceID, spaceID, c.Param("labelId")); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// AddToTicket はチケットにラベルを付ける（編集権限が要る）。
func (h *TicketLabelHandler) AddToTicket(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	if err := h.addToTicket.Execute(c.Request.Context(), ticket.AddTicketLabelInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, LabelID: c.Param("labelId"),
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveFromTicket はチケットからラベルを外す（編集権限が要る）。
func (h *TicketLabelHandler) RemoveFromTicket(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	if err := h.removeFrom.Execute(c.Request.Context(), ticket.RemoveTicketLabelInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, LabelID: c.Param("labelId"),
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
