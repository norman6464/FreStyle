package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// KnowledgeBaseGrantHandler はノートの「既定の権限（grant）」の読み書きを受ける。
//
// 認可はすべて kbPermissionGate が持つ（このファイルには判定規則を書かない）。
// なぜ handler 側で判定するのか / なぜ super_admin を特別扱いしないのか /
// なぜ拒否を 404 で揃えるのかは kb_permission_gate.go の冒頭を参照。
type KnowledgeBaseGrantHandler struct {
	*kbPermissionGate
	grantWorkspaceRole  *usecase.GrantWorkspaceRoleUseCase
	revokeWorkspaceRole *usecase.RevokeWorkspaceRoleUseCase
	grantSpaceRole      *usecase.GrantSpaceRoleUseCase
	revokeSpaceRole     *usecase.RevokeSpaceRoleUseCase
	grantPageRole       *usecase.GrantPageRoleUseCase
	revokePageRole      *usecase.RevokePageRoleUseCase
	listPageGrants      *usecase.ListPageGrantsUseCase
	listPrincipals      *usecase.ListGrantablePrincipalsUseCase
	canRemoveAdmin      *usecase.CanRemoveWorkspaceAdminUseCase
}

// NewKnowledgeBaseGrantHandler は KnowledgeBaseGrantHandler を組み立てる。
func NewKnowledgeBaseGrantHandler(
	gate *kbPermissionGate,
	grantWorkspaceRole *usecase.GrantWorkspaceRoleUseCase,
	revokeWorkspaceRole *usecase.RevokeWorkspaceRoleUseCase,
	grantSpaceRole *usecase.GrantSpaceRoleUseCase,
	revokeSpaceRole *usecase.RevokeSpaceRoleUseCase,
	grantPageRole *usecase.GrantPageRoleUseCase,
	revokePageRole *usecase.RevokePageRoleUseCase,
	listPageGrants *usecase.ListPageGrantsUseCase,
	listPrincipals *usecase.ListGrantablePrincipalsUseCase,
	canRemoveAdmin *usecase.CanRemoveWorkspaceAdminUseCase,
) *KnowledgeBaseGrantHandler {
	return &KnowledgeBaseGrantHandler{
		kbPermissionGate:    gate,
		grantWorkspaceRole:  grantWorkspaceRole,
		revokeWorkspaceRole: revokeWorkspaceRole,
		grantSpaceRole:      grantSpaceRole,
		revokeSpaceRole:     revokeSpaceRole,
		grantPageRole:       grantPageRole,
		revokePageRole:      revokePageRole,
		listPageGrants:      listPageGrants,
		listPrincipals:      listPrincipals,
		canRemoveAdmin:      canRemoveAdmin,
	}
}

// kbWorkspaceGrantResponse はワークスペース全体の既定の役割 1 件の返却形。
// workspaceId は載せない（URL の slug で決まるサーバ側の関心事）。
type kbWorkspaceGrantResponse struct {
	PrincipalID string    `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	Role        string    `json:"role"        example:"editor"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toKbWorkspaceGrantResponse(g *domain.WorkspaceGrant) kbWorkspaceGrantResponse {
	return kbWorkspaceGrantResponse{
		PrincipalID: g.PrincipalID,
		Role:        string(g.Role),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// kbSpaceGrantResponse はスペースの既定の役割 1 件の返却形。
type kbSpaceGrantResponse struct {
	SpaceID     string    `json:"spaceId"     example:"0198a000-0000-7000-8000-000000000002"`
	PrincipalID string    `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	Role        string    `json:"role"        example:"editor"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toKbSpaceGrantResponse(g *domain.SpaceGrant) kbSpaceGrantResponse {
	return kbSpaceGrantResponse{
		SpaceID:     g.SpaceID,
		PrincipalID: g.PrincipalID,
		Role:        string(g.Role),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// kbPageGrantResponse はページの既定の役割 1 件の返却形。
type kbPageGrantResponse struct {
	PageID      string    `json:"pageId"      example:"0198a000-0000-7000-8000-000000000003"`
	PrincipalID string    `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	Role        string    `json:"role"        example:"editor"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toKbPageGrantResponse(g *domain.PageGrant) kbPageGrantResponse {
	return kbPageGrantResponse{
		PageID:      g.PageID,
		PrincipalID: g.PrincipalID,
		Role:        string(g.Role),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

// kbGrantRoleRequest は grant を張るときの入力。
type kbGrantRoleRequest struct {
	// Role は domain.ValidGrantRoles のいずれか。既知でない値は usecase が弾く。
	Role string `json:"role" binding:"required" example:"editor"`
}

// requireNotLastWorkspaceAdmin は「最後の admin を剥がす操作」を断る。
// 断るときは応答を書いて false を返す。
//
// # なぜ剥がせなくしたか
//
// ノートの権限は principals / grants だけで閉じていて、
// アプリの super_admin による救済経路を意図的に持たない。ワークスペースの admin が
// 0 人になると、そこから先は権限を張り直す手段が API に存在せず、
// DB を直接触る以外に復旧できない。
//
// 反対に「最後の 1 人は自分を外せない」で困る場面は、先に別の誰かへ admin を渡せば
// 必ず解ける。取り返しがつかない側を禁じ、手数が 1 つ増えるだけの側を許す。
//
// 判定そのもの（何を admin として数えるか）は CanRemoveWorkspaceAdminUseCase の doc にある。
//
// **この検査は書き込みより手前の読み取りなので、これ単体では競合を防げない。**
// 同時に 2 人の admin を外す要求は両方ともここを通り抜け得る。実際に 0 人を止めているのは
// repository 側（判定と書き換えを同じトランザクションに入れ、admin の行を FOR UPDATE で
// ロックしてから決める）で、そこで断られた場合も respondKbPermissionOperationErr が
// 同じ 409 に落とす。ここは「日常の誤操作を、書き換えを試みる前に断る」ための層。
//
// 応答は 409（既にアーカイブ済み・循環と同じ「要求は正しいが対象の現在の状態と
// 両立しない」）。ここへ来る相手は admin なので、理由を返してよい
// （拒否を 404 に揃える規則は「権限が無い相手に対象を明かさない」ためのもので、
// admin 自身への説明までは縛らない）。
func (h *KnowledgeBaseGrantHandler) requireNotLastWorkspaceAdmin(
	c *gin.Context, in usecase.CanRemoveWorkspaceAdminInput,
) bool {
	ok, err := h.canRemoveAdmin.Execute(c.Request.Context(), in)
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return false
	}
	if !ok {
		c.JSON(http.StatusConflict, errorResponse{Error: "last_workspace_admin"})
		return false
	}
	return true
}

// GrantWorkspaceRole はワークスペース全体での既定の役割を主体に与える。
func (h *KnowledgeBaseGrantHandler) GrantWorkspaceRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	// 認可が先。落ちた要求は principalId にもボディにも触れない（対象の実在が漏れない）。
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbGrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	principalID := c.Param("principalId")
	// admin から他の役割へ落とすのも「admin を外す」操作。取り消しと同じ検査を通す。
	if domain.GrantRole(req.Role) != domain.GrantRoleAdmin {
		if !h.requireNotLastWorkspaceAdmin(c, usecase.CanRemoveWorkspaceAdminInput{
			WorkspaceID: scope.workspaceID,
			PrincipalID: principalID,
		}) {
			return
		}
	}
	grant, err := h.grantWorkspaceRole.Execute(c.Request.Context(), usecase.GrantWorkspaceRoleInput{
		WorkspaceID: scope.workspaceID,
		PrincipalID: principalID,
		Role:        domain.GrantRole(req.Role),
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbWorkspaceGrantResponse(grant))
}

// RevokeWorkspaceRole はワークスペース全体での既定の役割を剥がす（冪等）。
func (h *KnowledgeBaseGrantHandler) RevokeWorkspaceRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	principalID := c.Param("principalId")
	if !h.requireNotLastWorkspaceAdmin(c, usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: scope.workspaceID,
		PrincipalID: principalID,
	}) {
		return
	}
	if err := h.revokeWorkspaceRole.Execute(c.Request.Context(), usecase.RevokeWorkspaceRoleInput{
		WorkspaceID: scope.workspaceID,
		PrincipalID: principalID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GrantSpaceRole はスペースでの既定の役割を主体に与える。
func (h *KnowledgeBaseGrantHandler) GrantSpaceRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpaceAdmin(c, scope, spaceID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbGrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// 「最後の admin」の検査はここでは行わない。守っているのはワークスペースの admin が
	// 0 人になることで、スペースの admin を外してもワークスペースの admin は残る
	// （スペースの grant を全部消しても、ワークスペースの admin は配下の全スペースに届く）。
	grant, err := h.grantSpaceRole.Execute(c.Request.Context(), usecase.GrantSpaceRoleInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		PrincipalID: c.Param("principalId"),
		Role:        domain.GrantRole(req.Role),
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbSpaceGrantResponse(grant))
}

// RevokeSpaceRole はスペースでの既定の役割を剥がす（冪等）。
func (h *KnowledgeBaseGrantHandler) RevokeSpaceRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpaceAdmin(c, scope, spaceID) {
		return
	}
	if err := h.revokeSpaceRole.Execute(c.Request.Context(), usecase.RevokeSpaceRoleInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
		PrincipalID: c.Param("principalId"),
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GrantPageRole はページでの既定の役割を主体に与える。
func (h *KnowledgeBaseGrantHandler) GrantPageRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbGrantRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	// 「最後の admin」の検査はここでも行わない。守っているのはワークスペースの admin が
	// 0 人になることで、ページの grant をどう変えてもワークスペースの admin は
	// 配下の全ページに届き続ける（RevokeSpaceRole と同じ理由）。
	grant, err := h.grantPageRole.Execute(c.Request.Context(), usecase.GrantPageRoleInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		PrincipalID: c.Param("principalId"),
		Role:        domain.GrantRole(req.Role),
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageGrantResponse(grant))
}

// RevokePageRole はページでの既定の役割を剥がす（冪等）。
func (h *KnowledgeBaseGrantHandler) RevokePageRole(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	if err := h.revokePageRole.Execute(c.Request.Context(), usecase.RevokePageRoleInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		PrincipalID: c.Param("principalId"),
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPageGrants はそのページ自身に張られた既定の役割の一覧を返す。
func (h *KnowledgeBaseGrantHandler) ListPageGrants(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if !h.requirePageAdmin(c, scope, pageID) {
		return
	}
	grants, err := h.listPageGrants.Execute(c.Request.Context(), usecase.ListPageGrantsInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	out := make([]kbPageGrantResponse, 0, len(grants))
	for i := range grants {
		out = append(out, toKbPageGrantResponse(&grants[i]))
	}
	c.JSON(http.StatusOK, out)
}

// kbGrantablePrincipalResponse は権限を張れる相手 1 件の返却形。
type kbGrantablePrincipalResponse struct {
	ID   string `json:"id"   example:"0198a000-0000-7000-8000-00000000000a"`
	Kind string `json:"kind" example:"user"`
	// Name は表示名。引けなかった場合は空文字（行は落とさない）。
	Name string `json:"name" example:"田中 太郎"`
}

// ListGrantablePrincipals はそのページに権限を張れる相手を表示名つきで返す。
func (h *KnowledgeBaseGrantHandler) ListGrantablePrincipals(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	// 認可をページ単位で掛けるのは、この一覧を使うのが「そのページの権限を変えられる人」
	// だから。ワークスペースの admin に絞ると、ページに admin を張られた人が
	// 相手を選べなくなる（権限はあるのに画面が使えない）。
	if !h.requirePageAdmin(c, scope, c.Param("pageId")) {
		return
	}
	principals, err := h.listPrincipals.Execute(c.Request.Context(), usecase.ListGrantablePrincipalsInput{
		WorkspaceID: scope.workspaceID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	out := make([]kbGrantablePrincipalResponse, 0, len(principals))
	for _, p := range principals {
		out = append(out, kbGrantablePrincipalResponse{ID: p.ID, Kind: string(p.Kind), Name: p.Name})
	}
	c.JSON(http.StatusOK, out)
}
