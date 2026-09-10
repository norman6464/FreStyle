package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/kb"
)

// PageTemplateHandler はページの雛形（page_templates）操作を受ける。
//
// 認可の方針: 雛形はワークスペース全体で共有される資産なので、作成・削除は
// ワークスペースレベルの CanEdit（CheckWorkspacePermissionUseCase 経由）で判定する
// （CreateSpace の CanManage 判定と同じパターンで、揃える段だけ CanEdit にする）。
// 一覧はワークスペース所属者なら誰でも読める（IsWorkspaceMemberUseCase）。
// 「雛形から作る」（実際にページを作る操作）は既存のページ作成（KnowledgeBasePageHandler.Create）
// と全く同じ認可分岐——親の有無で requireSpacePermission / requirePagePermission を切り替える。
//
// ここまでは handler の役割。**特定のスペースに限定した雛形**（PageTemplate.SpaceID != nil）
// については、この handler の判定だけでは足りない — ワークスペース全体への CanEdit や
// 作成先の場所への権限は、雛形自身がひも付く非公開スペースを見てよいかとは別物のため、
// 一覧の spaceId 絞り込み・作成・削除・使用のそれぞれで usecase 側が対象スペースへの
// CanView を追加で確かめる（kb.ListPageTemplatesUseCase 等の doc コメント参照）。
type PageTemplateHandler struct {
	isMember       *kb.IsWorkspaceMemberUseCase
	checkWorkspace *kb.CheckWorkspacePermissionUseCase
	checkPage      *kb.CheckPagePermissionUseCase
	checkSpace     *kb.CheckSpacePermissionUseCase
	list           *kb.ListPageTemplatesUseCase
	createTemplate *kb.CreateTemplateFromPageUseCase
	deleteTemplate *kb.DeletePageTemplateUseCase
	createPage     *kb.CreatePageFromTemplateUseCase
}

// NewPageTemplateHandler は PageTemplateHandler を組み立てる。
func NewPageTemplateHandler(
	isMember *kb.IsWorkspaceMemberUseCase,
	checkWorkspace *kb.CheckWorkspacePermissionUseCase,
	checkPage *kb.CheckPagePermissionUseCase,
	checkSpace *kb.CheckSpacePermissionUseCase,
	list *kb.ListPageTemplatesUseCase,
	createTemplate *kb.CreateTemplateFromPageUseCase,
	deleteTemplate *kb.DeletePageTemplateUseCase,
	createPage *kb.CreatePageFromTemplateUseCase,
) *PageTemplateHandler {
	return &PageTemplateHandler{
		isMember:       isMember,
		checkWorkspace: checkWorkspace,
		checkPage:      checkPage,
		checkSpace:     checkSpace,
		list:           list,
		createTemplate: createTemplate,
		deleteTemplate: deleteTemplate,
		createPage:     createPage,
	}
}

// kbPageTemplateResponse は雛形 1 件の返却形。doc は含まない（一覧・作成の応答を軽量に保つ。
// 使うときは CreatePage の応答が既存の kbPageResponse と同じ形で返るので、フロントは
// 別途 GET で本文を取りに行く前提——KnowledgeBasePageHandler.Create と同じ応答契約）。
type kbPageTemplateResponse struct {
	ID        string              `json:"id"                example:"0198a000-0000-7000-8000-000000000010"`
	Name      string              `json:"name"               example:"議事録"`
	Icon      *kbPageIconResponse `json:"icon,omitempty"`
	SpaceID   *string             `json:"spaceId,omitempty"  example:"0198a000-0000-7000-8000-000000000002"`
	CreatedAt time.Time           `json:"createdAt"`
}

func toKbPageTemplateResponse(t *domain.PageTemplate) kbPageTemplateResponse {
	resp := kbPageTemplateResponse{ID: t.ID, Name: t.Name, SpaceID: t.SpaceID, CreatedAt: t.CreatedAt}
	if t.Icon != nil {
		resp.Icon = &kbPageIconResponse{Type: string(t.Icon.Type), Value: t.Icon.Value}
	}
	return resp
}

// requireWorkspaceMember はワークスペース所属者であることだけを確かめる（雛形の一覧はこれで
// 足りる。編集の役割の強弱は問わない）。
func (h *PageTemplateHandler) requireWorkspaceMember(c *gin.Context, scope kbRequestScope) bool {
	ok, err := h.isMember.Execute(c.Request.Context(), kb.IsWorkspaceMemberInput{
		WorkspaceID: scope.workspaceID, UserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return false
	}
	if !ok {
		// ここに来る相手は middleware.KnowledgeBaseWorkspace を通過済み（= 所属者）が
		// 前提のため、通常は起こらない。フェイルクローズ側に倒す。
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// requireWorkspaceCanEdit はワークスペース全体への書き込み資格（CanEdit）を確かめる。
// 雛形の作成・削除はこの 1 段だけで判定する（CreateSpace の CanManage 判定と同じ形）。
func (h *PageTemplateHandler) requireWorkspaceCanEdit(c *gin.Context, scope kbRequestScope) bool {
	perm, err := h.checkWorkspace.Execute(c.Request.Context(), kb.CheckWorkspacePermissionInput{
		WorkspaceID: scope.workspaceID, UserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return false
	}
	if !perm.CanEdit {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return false
	}
	return true
}

// List はワークスペース（または spaceId クエリで絞った特定のスペース）の雛形一覧を返す。
// ワークスペース所属者なら誰でも読める。
func (h *PageTemplateHandler) List(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceMember(c, scope) {
		return
	}
	var spaceID *string
	if q := c.Query("spaceId"); q != "" {
		spaceID = &q
	}
	templates, err := h.list.Execute(c.Request.Context(), kb.ListPageTemplatesInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		UserID:      scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	// 0 件でも [] を返す（null だとフロントの .map が落ちる）。
	out := make([]kbPageTemplateResponse, 0, len(templates))
	for i := range templates {
		out = append(out, toKbPageTemplateResponse(&templates[i]))
	}
	c.JSON(http.StatusOK, out)
}

// kbCreateTemplateRequest は「雛形として保存」の入力。
type kbCreateTemplateRequest struct {
	Name string `json:"name" binding:"required" example:"議事録"`
	// SpaceID が nil ならワークスペース全体で見える雛形になる。
	SpaceID *string `json:"spaceId,omitempty" example:"0198a000-0000-7000-8000-000000000002"`
}

// CreateFromPage はページの今の本文を雛形として保存する。
// そのページ自体を編集できること（requirePagePermission）と、ワークスペース全体への
// 書き込み資格（CanEdit）の両方を満たさないと 403 になる。
func (h *PageTemplateHandler) CreateFromPage(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !requirePagePermissionWith(c, h.checkPage, scope, pageID, domain.CapabilityEdit) {
		return
	}
	if !h.requireWorkspaceCanEdit(c, scope) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	tpl, err := h.createTemplate.Execute(c.Request.Context(), kb.CreateTemplateFromPageInput{
		WorkspaceID:  scope.workspaceID,
		PageID:       pageID,
		SpaceID:      req.SpaceID,
		Name:         req.Name,
		AuthorUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbPageTemplateResponse(tpl))
}

// Delete は雛形を削除する（ワークスペース全体への CanEdit のみで判定する）。
func (h *PageTemplateHandler) Delete(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceCanEdit(c, scope) {
		return
	}
	templateID := c.Param("templateId")
	if err := h.deleteTemplate.Execute(c.Request.Context(), kb.DeletePageTemplateInput{
		WorkspaceID: scope.workspaceID,
		TemplateID:  templateID,
		UserID:      scope.userID,
	}); err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// kbCreatePageFromTemplateRequest は「雛形から作る」の入力。
type kbCreatePageFromTemplateRequest struct {
	TemplateID string `json:"templateId" binding:"required"`
	// ParentID が空文字（未指定）ならスペース直下に作る（kbCreatePageRequest と同じ意味）。
	ParentID string `json:"parentId,omitempty" example:"0198a000-0000-7000-8000-000000000003"`
	Title    string `json:"title"    binding:"required,max=200" example:"設計メモ"`
}

// CreatePage は雛形から新しいページを作る。認可分岐は既存の Create（kb_page_handler.go）と
// 全く同じ——parentId が空ならそのスペースの編集権限、あればその親ページの編集権限を見る
// （雛形そのものへの権限は問わない。テンプレート一覧を読めた時点でどの雛形かは分かっている）。
func (h *PageTemplateHandler) CreatePage(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreatePageFromTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	spaceID := c.Param("spaceId")
	var parentID *string
	if req.ParentID == "" {
		if !requireSpacePermissionWith(c, h.checkSpace, scope, spaceID, domain.CapabilityEdit) {
			return
		}
	} else {
		if !requirePagePermissionWith(c, h.checkPage, scope, req.ParentID, domain.CapabilityEdit) {
			return
		}
		parentID = &req.ParentID
	}
	out, err := h.createPage.Execute(c.Request.Context(), kb.CreatePageFromTemplateInput{
		WorkspaceID:  scope.workspaceID,
		SpaceID:      spaceID,
		ParentID:     parentID,
		TemplateID:   req.TemplateID,
		Title:        req.Title,
		AuthorUserID: scope.userID,
	})
	if err != nil {
		respondKnowledgeBaseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbPageResponse(&out.Page))
}
