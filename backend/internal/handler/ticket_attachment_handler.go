package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/ticket"
)

// TicketAttachmentHandler はチケットへのファイル添付を受ける（段 4）。
// アップロードは 2 段（1: presigned PUT URL の発行 → クライアントが直接 Cloud Storage へ
// PUT → 2: メタデータの記録）。kb ページ画像・rich-text 画像はどちらもアップロードのみで
// 「記録」の概念を持たない（本文 JSON に直接 key を埋め込むだけ）が、添付は一覧・削除の
// 対象になる独立した「もの」なので、記録用の表 1 つを別に持つ（設計 Ⅵ 着手時の結論）。
type TicketAttachmentHandler struct {
	checkTicket   *ticket.CheckTicketPermissionUseCase
	issueUpload   *ticket.IssueTicketAttachmentUploadURLUseCase
	create        *ticket.CreateTicketAttachmentUseCase
	list          *ticket.ListTicketAttachmentsUseCase
	issueDownload *ticket.IssueTicketAttachmentDownloadURLUseCase
	del           *ticket.DeleteTicketAttachmentUseCase
}

func NewTicketAttachmentHandler(
	checkTicket *ticket.CheckTicketPermissionUseCase,
	issueUpload *ticket.IssueTicketAttachmentUploadURLUseCase,
	create *ticket.CreateTicketAttachmentUseCase,
	list *ticket.ListTicketAttachmentsUseCase,
	issueDownload *ticket.IssueTicketAttachmentDownloadURLUseCase,
	del *ticket.DeleteTicketAttachmentUseCase,
) *TicketAttachmentHandler {
	return &TicketAttachmentHandler{
		checkTicket: checkTicket, issueUpload: issueUpload, create: create,
		list: list, issueDownload: issueDownload, del: del,
	}
}

func (h *TicketAttachmentHandler) requireTicketPermission(
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

// ticketAttachmentListResponse は添付一覧の返却形。
type ticketAttachmentListResponse struct {
	Attachments []domain.TicketAttachment `json:"attachments"`
}

// List はチケットの添付一覧を返す（閲覧権限が要る）。
func (h *TicketAttachmentHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityView) {
		return
	}
	attachments, err := h.list.Execute(c.Request.Context(), scope.workspaceID, ticketID)
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	if attachments == nil {
		attachments = []domain.TicketAttachment{}
	}
	c.JSON(http.StatusOK, ticketAttachmentListResponse{Attachments: attachments})
}

// ticketAttachmentUploadURLRequest はアップロード URL 発行の入力。
type ticketAttachmentUploadURLRequest struct {
	ContentType string `json:"contentType" binding:"required"`
	Size        int64  `json:"size" binding:"required"`
}

// ticketAttachmentUploadURLResponse はアップロード URL 発行の応答形。
type ticketAttachmentUploadURLResponse struct {
	URL       string `json:"url"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}

// IssueUploadURL は添付ファイルの PUT presigned URL を発行する（編集権限が要る）。
func (h *TicketAttachmentHandler) IssueUploadURL(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	var req ticketAttachmentUploadURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	out, err := h.issueUpload.Execute(c.Request.Context(), ticket.IssueTicketAttachmentUploadURLInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, ContentType: req.ContentType, Size: req.Size,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticketAttachmentUploadURLResponse{URL: out.URL, Key: out.Key, ExpiresIn: out.ExpiresIn})
}

// ticketAttachmentCreateRequest は添付メタデータ記録の入力（アップロード URL 発行の応答が
// 返した key をそのまま渡す）。
type ticketAttachmentCreateRequest struct {
	Key         string `json:"key" binding:"required"`
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
	SizeBytes   int64  `json:"sizeBytes" binding:"required"`
}

// Create はクライアントが PUT を終えたあとに添付のメタデータを記録する（編集権限が要る）。
func (h *TicketAttachmentHandler) Create(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	var req ticketAttachmentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	a, err := h.create.Execute(c.Request.Context(), ticket.CreateTicketAttachmentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, Key: req.Key, Filename: req.Filename,
		ContentType: req.ContentType, SizeBytes: req.SizeBytes, UploadedByUserID: scope.userID,
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, a)
}

// ticketAttachmentDownloadURLResponse はダウンロード URL 発行の応答形。
type ticketAttachmentDownloadURLResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expiresIn"`
}

// IssueDownloadURL は添付ファイルの GET（ダウンロード）presigned URL を発行する（閲覧権限が要る）。
func (h *TicketAttachmentHandler) IssueDownloadURL(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityView) {
		return
	}
	out, err := h.issueDownload.Execute(c.Request.Context(), ticket.IssueTicketAttachmentDownloadURLInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, AttachmentID: c.Param("attachmentId"),
	})
	if err != nil {
		respondTicketErr(c, err)
		return
	}
	c.JSON(http.StatusOK, ticketAttachmentDownloadURLResponse{URL: out.URL, ExpiresIn: out.ExpiresIn})
}

// Delete は添付を削除する（編集権限が要る）。
func (h *TicketAttachmentHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	ticketID := c.Param("ticketId")
	if !h.requireTicketPermission(c, scope, ticketID, domain.CapabilityEdit) {
		return
	}
	if err := h.del.Execute(c.Request.Context(), ticket.DeleteTicketAttachmentInput{
		WorkspaceID: scope.workspaceID, TicketID: ticketID, AttachmentID: c.Param("attachmentId"),
	}); err != nil {
		respondTicketErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
