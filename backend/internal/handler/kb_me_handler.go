package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/handler/middleware"
	"github.com/norman6464/frestyle/backend/internal/usecase/kb"
)

// KnowledgeBaseMeHandler はワークスペースをまたぐ「自分」向けのナレッジ操作を受ける。
// URL に workspaceSlug を持たないため middleware.KnowledgeBaseWorkspace を通れない
// （所属先を一つに決めない操作専用。resolve/invitations と同じ理由で g に直接登録する）。
type KnowledgeBaseMeHandler struct {
	listRecentPages *kb.ListMyRecentPagesUseCase
}

// NewKnowledgeBaseMeHandler は KnowledgeBaseMeHandler を組み立てる。
func NewKnowledgeBaseMeHandler(listRecentPages *kb.ListMyRecentPagesUseCase) *KnowledgeBaseMeHandler {
	return &KnowledgeBaseMeHandler{listRecentPages: listRecentPages}
}

// kbRecentPageResponse は「最近見たページ」1 件の返却形。workspaceId は載せない
// （kbPageResponse と同じ方針 — クライアントは workspaceSlug で以後の API を呼ぶ）。
type kbRecentPageResponse struct {
	PageID        string              `json:"pageId"`
	WorkspaceSlug string              `json:"workspaceSlug"`
	Title         string              `json:"title"`
	Icon          *kbPageIconResponse `json:"icon,omitempty"`
	SpaceID       string              `json:"spaceId"`
	SpaceName     string              `json:"spaceName"`
	ViewedAt      string              `json:"viewedAt"`
}

func toKbRecentPageResponse(p domain.RecentPage) kbRecentPageResponse {
	resp := kbRecentPageResponse{
		PageID:        p.PageID,
		WorkspaceSlug: p.WorkspaceSlug,
		Title:         p.Title,
		SpaceID:       p.SpaceID,
		SpaceName:     p.SpaceName,
		ViewedAt:      p.ViewedAt.Format("2006-01-02T15:04:05.000Z07:00"),
	}
	if p.Icon != nil {
		resp.Icon = &kbPageIconResponse{Type: string(p.Icon.Type), Value: p.Icon.Value}
	}
	return resp
}

// ListRecentPages は自分が最近見たページを新しい順に返す（ワークスペース横断）。
func (h *KnowledgeBaseMeHandler) ListRecentPages(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		return
	}
	pages, err := h.listRecentPages.Execute(c.Request.Context(), uid)
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	out := make([]kbRecentPageResponse, 0, len(pages))
	for _, p := range pages {
		out = append(out, toKbRecentPageResponse(p))
	}
	c.JSON(http.StatusOK, out)
}
