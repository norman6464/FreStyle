package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// KnowledgeBaseGrantHandler はナレッジ基盤の「既定の権限（grant）」と
// 「ページ単位の例外（restriction）」の書き換えを受ける。
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
	setRestriction      *usecase.SetPageRestrictionUseCase
	clearRestriction    *usecase.ClearPageRestrictionUseCase
	canRemoveAdmin      *usecase.CanRemoveWorkspaceAdminUseCase
}

// NewKnowledgeBaseGrantHandler は KnowledgeBaseGrantHandler を組み立てる。
func NewKnowledgeBaseGrantHandler(
	gate *kbPermissionGate,
	grantWorkspaceRole *usecase.GrantWorkspaceRoleUseCase,
	revokeWorkspaceRole *usecase.RevokeWorkspaceRoleUseCase,
	grantSpaceRole *usecase.GrantSpaceRoleUseCase,
	revokeSpaceRole *usecase.RevokeSpaceRoleUseCase,
	setRestriction *usecase.SetPageRestrictionUseCase,
	clearRestriction *usecase.ClearPageRestrictionUseCase,
	canRemoveAdmin *usecase.CanRemoveWorkspaceAdminUseCase,
) *KnowledgeBaseGrantHandler {
	return &KnowledgeBaseGrantHandler{
		kbPermissionGate:    gate,
		grantWorkspaceRole:  grantWorkspaceRole,
		revokeWorkspaceRole: revokeWorkspaceRole,
		grantSpaceRole:      grantSpaceRole,
		revokeSpaceRole:     revokeSpaceRole,
		setRestriction:      setRestriction,
		clearRestriction:    clearRestriction,
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

// kbPageRestrictionResponse はページ単位の例外 1 件の返却形。
type kbPageRestrictionResponse struct {
	PageID      string    `json:"pageId"      example:"0198a000-0000-7000-8000-000000000003"`
	PrincipalID string    `json:"principalId" example:"0198a000-0000-7000-8000-00000000000a"`
	Capability  string    `json:"capability"  example:"view"`
	Mode        string    `json:"mode"        example:"deny"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func toKbPageRestrictionResponse(r *domain.PageRestriction) kbPageRestrictionResponse {
	return kbPageRestrictionResponse{
		PageID:      r.PageID,
		PrincipalID: r.PrincipalID,
		Capability:  string(r.Capability),
		Mode:        string(r.Mode),
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// kbGrantRoleRequest は grant を張るときの入力。
type kbGrantRoleRequest struct {
	// Role は domain.ValidGrantRoles のいずれか。既知でない値は usecase が弾く。
	Role string `json:"role" binding:"required" example:"editor"`
}

// kbSetRestrictionRequest はページの例外を設定するときの入力。
type kbSetRestrictionRequest struct {
	// Mode は allow（限定公開の許可リストへ載せる）か deny（この主体だけ外す）。
	Mode string `json:"mode" binding:"required" example:"deny"`
}

// requireNotLastWorkspaceAdmin は「最後の admin を剥がす操作」を断る。
// 断るときは応答を書いて false を返す。
//
// # なぜ剥がせなくしたか
//
// ナレッジ基盤の権限は principals / grants / restrictions だけで閉じていて、
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
//	@Summary      ナレッジ 基盤 の ワークスペース 権限 付与
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
//	@Summary      ナレッジ 基盤 の ワークスペース 権限 取り消し
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
//	@Summary      ナレッジ 基盤 の スペース 権限 付与
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
//	@Summary      ナレッジ 基盤 の スペース 権限 取り消し
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

// SetPageRestriction はページ以下だけ既定を上書きする例外を設定する。
//
//	@Summary      ナレッジ 基盤 の ページ 例外 設定
//	@Description  ページ と その 子孫 に だけ 効く 例外 を 1 行 設定 する。 mode=deny は 名指し し た 主体 だけ を 外す (ほか の 人 の 既定 は 変わら ない)。 mode=allow は その ページ の その ケイパビリティ を 「載っ て いる 主体 だけ」 の 限定 公開 に 切り替える。 呼べる の は その ページ が 属する スペース の admin (ワークスペース の admin を 含む) だけ。 閲覧 権限 は 要求 し ない (自分 を deny し た ページ の 例外 を 自分 で 戻せ なく なる ため)。 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                   true  "ワークスペース の slug"
//	@Param        pageId         path      string                   true  "ページ ID (UUID)"
//	@Param        principalId    path      string                   true  "主体 ID (UUID)"
//	@Param        capability     path      string                   true  "ケイパビリティ (view / edit)"
//	@Param        body           body      kbSetRestrictionRequest  true  "例外 の 向き (allow / deny)"
//	@Success      200            {object}  kbPageRestrictionResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/restrictions/{principalId}/{capability} [put]
//	@Security     CookieAuth
func (h *KnowledgeBaseGrantHandler) SetPageRestriction(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if _, ok := h.requirePageAdmin(c, scope, pageID); !ok {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbSetRestrictionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	restriction, err := h.setRestriction.Execute(c.Request.Context(), usecase.SetPageRestrictionInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		PrincipalID: c.Param("principalId"),
		Capability:  domain.Capability(c.Param("capability")),
		Mode:        domain.RestrictionMode(req.Mode),
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPageRestrictionResponse(restriction))
}

// ClearPageRestriction はページの例外を 1 行解除する（冪等）。
//
//	@Summary      ナレッジ 基盤 の ページ 例外 解除
//	@Description  ページ に 張ら れ た 例外 を 1 行 解除 する。 元 から 無い 行 に 対し て も 成功 する (冪等)。 消し た の が 最後 の allow 行 なら 限定 公開 も 畳ま れ、 解決 は より 遠い 祖先 の 制限 → grant の 既定 へ 戻る。 deny 行 の 解除 で は 限定 公開 を 畳ま ない。 呼べる の は その ページ が 属する スペース の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string  true  "ワークスペース の slug"
//	@Param        pageId         path  string  true  "ページ ID (UUID)"
//	@Param        principalId    path  string  true  "主体 ID (UUID)"
//	@Param        capability     path  string  true  "ケイパビリティ (view / edit)"
//	@Success      204            "解除 済み"
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/pages/{pageId}/restrictions/{principalId}/{capability} [delete]
//	@Security     CookieAuth
func (h *KnowledgeBaseGrantHandler) ClearPageRestriction(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	pageID := c.Param("pageId")
	if _, ok := h.requirePageAdmin(c, scope, pageID); !ok {
		return
	}
	if err := h.clearRestriction.Execute(c.Request.Context(), usecase.ClearPageRestrictionInput{
		WorkspaceID: scope.workspaceID,
		PageID:      pageID,
		PrincipalID: c.Param("principalId"),
		Capability:  domain.Capability(c.Param("capability")),
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
