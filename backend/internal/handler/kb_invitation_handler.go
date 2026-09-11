package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// KnowledgeBaseInvitationHandler は「自分宛の招待」を受ける。
//
// メンバー追加（KnowledgeBaseMemberHandler.InviteMember）は workspace_members に
// invited の行を作るだけで、principal・権限はこの handler の Accept を本人が呼ぶまで
// 発生しない（段 2。FRESTYLE-486 の修正）。まだ所属していない状態で叩くエンドポイントなので
// middleware.KnowledgeBaseWorkspace（所属済みしか通さない）を通さない
// （KnowledgeBaseWorkspaceHandler.List / Create と同じ位置付け）。
type KnowledgeBaseInvitationHandler struct {
	list    *kb.ListMyWorkspaceInvitationsUseCase
	accept  *kb.AcceptWorkspaceInvitationUseCase
	decline *kb.DeclineWorkspaceInvitationUseCase
}

// NewKnowledgeBaseInvitationHandler は KnowledgeBaseInvitationHandler を組み立てる。
func NewKnowledgeBaseInvitationHandler(
	list *kb.ListMyWorkspaceInvitationsUseCase,
	accept *kb.AcceptWorkspaceInvitationUseCase,
	decline *kb.DeclineWorkspaceInvitationUseCase,
) *KnowledgeBaseInvitationHandler {
	return &KnowledgeBaseInvitationHandler{list: list, accept: accept, decline: decline}
}

// kbInvitationResponse は招待 1 件の返却形。
type kbInvitationResponse struct {
	WorkspaceSlug   string    `json:"workspaceSlug" example:"acme"`
	WorkspaceName   string    `json:"workspaceName" example:"Acme 社"`
	InvitedByUserID uint64    `json:"invitedByUserId" example:"7"`
	InvitedAt       time.Time `json:"invitedAt"`
}

func toKbInvitationResponse(inv *domain.WorkspaceInvitation) kbInvitationResponse {
	return kbInvitationResponse{
		WorkspaceSlug:   inv.WorkspaceSlug,
		WorkspaceName:   inv.WorkspaceName,
		InvitedByUserID: inv.InvitedByUserID,
		InvitedAt:       inv.InvitedAt,
	}
}

// List は自分宛の未受諾の招待を返す。
func (h *KnowledgeBaseInvitationHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	invitations, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	out := make([]kbInvitationResponse, 0, len(invitations))
	for i := range invitations {
		out = append(out, toKbInvitationResponse(&invitations[i]))
	}
	c.JSON(http.StatusOK, out)
}

// Accept は自分宛の招待を受諾する。principal（kind='user'）を作り、既定の editor を与える。
func (h *KnowledgeBaseInvitationHandler) Accept(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	ws, err := h.accept.Execute(c.Request.Context(), kb.AcceptWorkspaceInvitationInput{
		WorkspaceSlug: c.Param("workspaceSlug"),
		UserID:        uid,
	})
	if err != nil {
		if errors.Is(err, repository.ErrWorkspaceInvitationNotFound) || errors.Is(err, repository.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
			return
		}
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbWorkspaceResponse(ws, false))
}

// Decline は自分宛の招待を辞退する。
func (h *KnowledgeBaseInvitationHandler) Decline(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	if err := h.decline.Execute(c.Request.Context(), kb.DeclineWorkspaceInvitationInput{
		WorkspaceSlug: c.Param("workspaceSlug"),
		UserID:        uid,
	}); err != nil {
		if errors.Is(err, repository.ErrWorkspaceInvitationNotFound) || errors.Is(err, repository.ErrWorkspaceNotFound) {
			c.JSON(http.StatusNotFound, errorResponse{Error: "not_found"})
			return
		}
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
