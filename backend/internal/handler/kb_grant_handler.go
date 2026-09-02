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
//
//	@Summary      ノート の ワークスペース 権限 付与
//	@Description  ワークスペース 全体 で の 既定 の 役割 を 主体 に 与える (同じ 主体 に は 1 行 だけ な の で 上書き)。 配下 の 全 スペース に 効く。 呼べる の は ワークスペース の admin だけ。 権限 が 無い 場合 と 対象 (ワークスペース / 主体) が 存在 し ない 場合 は、 実在 を 漏らさ ない よう 同じ 404 を 返す。 admin を 外す 向き の 変更 で、 ユーザー の admin が 1 人 も 残ら なく なる とき は 409。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string              true  "ワークスペース の slug"
//	@Param        principalId    path      string              true  "主体 ID (UUID)"
//	@Param        body           body      kbGrantRoleRequest  true  "役割 (admin / editor / commenter / viewer)"
//	@Success      200            {object}  kbWorkspaceGrantResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      409            {object}  errorResponse  "最後 の admin を 外す 操作"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/grants/{principalId} [put]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の ワークスペース 権限 取り消し
//	@Description  ワークスペース 全体 で の 既定 の 役割 を 剥がす。 元 から 無い 相手 に 対し て も 成功 する (冪等)。 呼べる の は ワークスペース の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。 ユーザー の admin が 1 人 も 残ら なく なる とき は 409 で 断る (誰 も 権限 を 変え られ ない ワークスペース は API から 復旧 でき ない ため)。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        principalId    path  string  true  "主体 ID (UUID)"
//	@Success      204            "取り消し 済み"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      409            {object}  errorResponse  "最後 の admin を 外す 操作"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/grants/{principalId} [delete]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の スペース 権限 付与
//	@Description  スペース で の 既定 の 役割 を 主体 に 与える (同じ 主体 に は 1 行 だけ)。 呼べる の は その スペース の admin (ワークスペース の admin を 含む) だけ。 スペース 単位 の grant で ワークスペース の admin が 降格 する こと は ない (最も 強い 役割 を 採る)。 権限 が 無い 場合 と 対象 (スペース / 主体) が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string              true  "ワークスペース の slug"
//	@Param        spaceId        path      string              true  "スペース ID (UUID)"
//	@Param        principalId    path      string              true  "主体 ID (UUID)"
//	@Param        body           body      kbGrantRoleRequest  true  "役割 (admin / editor / commenter / viewer)"
//	@Success      200            {object}  kbSpaceGrantResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/grants/{principalId} [put]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の スペース 権限 取り消し
//	@Description  スペース で の 既定 の 役割 を 剥がす。 元 から 無い 相手 に 対し て も 成功 する (冪等)。 呼べる の は その スペース の admin (ワークスペース の admin を 含む) だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        spaceId        path  string  true  "スペース ID (UUID)"
//	@Param        principalId    path  string  true  "主体 ID (UUID)"
//	@Success      204            "取り消し 済み"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/grants/{principalId} [delete]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の ページ 権限 付与
//	@Description  ページ で の 既定 の 役割 を 主体 に 与える (同じ 主体 に は 1 行 だけ)。 既定 の 3 段目 で、 この ページ と その 子孫 に 効く。 合成 は 上 の 2 段 と 同じ で、 複数 の 経路 から 届い た 役割 の うち 最も 強い もの が 実効 に なる ため、 **ここ で 誰か を 弱める こと は でき ない** (上位 で editor を 得 て いる 相手 に viewer を 張っ て も editor の まま)。 弱める 手段 は どの 層 に も 無い ので、 狭め たい 内容 は private の スペース へ 置く。 呼べる の は その ページ の admin (スペース / ワークスペース から 届い て いる 場合 を 含む) だけ。 権限 が 無い 場合 と 対象 (ページ / 主体) が 存在 し ない 場合 は、 実在 を 漏らさ ない よう 同じ 404 を 返す。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string              true  "ワークスペース の slug"
//	@Param        pageId         path      string              true  "ページ ID (UUID)"
//	@Param        principalId    path      string              true  "主体 ID (UUID)"
//	@Param        body           body      kbGrantRoleRequest  true  "役割 (admin / editor / commenter / viewer)"
//	@Success      200            {object}  kbPageGrantResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/grants/{principalId} [put]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の ページ 権限 取り消し
//	@Description  ページ で の 既定 の 役割 を 剥がす。 元 から 無い 相手 に 対し て も 成功 する (冪等)。 消える の は この 段 で 足し た 分 だけ で、 ワークスペース / スペース / 祖先 の ページ から 届い て いる 役割 は そのまま 残る (「この ページ だけ 見せ ない」 は 書け ない — 狭め たい 内容 は private の スペース へ 置く)。 呼べる の は その ページ の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Param        principalId    path  string  true  "主体 ID (UUID)"
//	@Success      204            "取り消し 済み"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/grants/{principalId} [delete]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の ページ 権限 一覧
//	@Description  その ページ 自身 に 張ら れ た 既定 の 役割 を 返す (祖先 から 降り て くる 分 は 含ま ない)。 **「この ページ を 見 られる 人 の 一覧」 で は ない** — ワークスペース / スペース の grant で 届い て いる 相手 も、 祖先 の ページ に 張ら れ た grant で 届い て いる 相手 も 含ま れ ない。 空 で 返っ て き て も 「誰 も 見 られ ない」 で は なく 「この 段 で は 何 も 足し て い ない」 の 意味。 呼べる の は その ページ の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Success      200            {array}   kbPageGrantResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/grants [get]
//	@Security     CookieAuth
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
//
//	@Summary      ノート の 権限 を 張れる 相手 の 一覧
//	@Description  その ページ に 権限 を 張れる 相手 (ユーザー / グループ / スペース の 全員) を 表示 名 つき で 返す。 共有 の 画面 で 相手 を 選ぶ ため の 口。 リンク の 来訪者 を 表す 主体 (share_link) は 含ま ない — あれ は リンク の 発行 時 に 自動 で 作ら れる もの で、 人 が 選ん で 役割 を 与える 相手 で は ない。 名前 が 引け なかっ た 行 も 空文字 の まま 返す (一覧 から 黙っ て 消す と、 その 相手 に 張っ た 権限 が 画面 に 残っ た まま 選べ なく なる)。 呼べる の は その ページ の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Success      200            {array}   kbGrantablePrincipalResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/principals [get]
//	@Security     CookieAuth
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
