package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// KnowledgeBaseMemberHandler はナレッジ基盤の主体（principals）の出し入れを受ける。
//
// ワークスペース所属・グループ・スペースの「全員」は、どれも principals の 1 行で表す
// （専用のメンバーシップ表は持たない）。したがってこの handler が扱うのは
// 「権限を張る相手を用意する / 片づける」ことで、役割そのものは
// KnowledgeBaseGrantHandler が扱う。
//
// 認可はすべて kbPermissionGate が持つ。判断の根拠は kb_permission_gate.go の冒頭を参照。
type KnowledgeBaseMemberHandler struct {
	*kbPermissionGate
	addMember         *usecase.AddWorkspaceMemberUseCase
	removeMember      *usecase.RemoveWorkspaceMemberUseCase
	createGroup       *usecase.CreatePrincipalGroupUseCase
	addGroupMember    *usecase.AddGroupMemberUseCase
	removeGroupMember *usecase.RemoveGroupMemberUseCase
	ensureEveryone    *usecase.EnsureSpaceEveryonePrincipalUseCase
	canRemoveAdmin    *usecase.CanRemoveWorkspaceAdminUseCase
}

// NewKnowledgeBaseMemberHandler は KnowledgeBaseMemberHandler を組み立てる。
func NewKnowledgeBaseMemberHandler(
	gate *kbPermissionGate,
	addMember *usecase.AddWorkspaceMemberUseCase,
	removeMember *usecase.RemoveWorkspaceMemberUseCase,
	createGroup *usecase.CreatePrincipalGroupUseCase,
	addGroupMember *usecase.AddGroupMemberUseCase,
	removeGroupMember *usecase.RemoveGroupMemberUseCase,
	ensureEveryone *usecase.EnsureSpaceEveryonePrincipalUseCase,
	canRemoveAdmin *usecase.CanRemoveWorkspaceAdminUseCase,
) *KnowledgeBaseMemberHandler {
	return &KnowledgeBaseMemberHandler{
		kbPermissionGate:  gate,
		addMember:         addMember,
		removeMember:      removeMember,
		createGroup:       createGroup,
		addGroupMember:    addGroupMember,
		removeGroupMember: removeGroupMember,
		ensureEveryone:    ensureEveryone,
		canRemoveAdmin:    canRemoveAdmin,
	}
}

// kbPrincipalResponse は主体 1 件の返却形。
//
// workspaceId は載せない（URL の slug で決まる）。id は載せる — grant / restriction を
// 張る URL がこの ID を取るので、クライアントが知る必要がある唯一の内部 ID。
type kbPrincipalResponse struct {
	ID   string `json:"id"   example:"0198a000-0000-7000-8000-00000000000a"`
	Kind string `json:"kind" example:"user"`
	// UserID は kind=user のときだけ入る。
	UserID *uint64 `json:"userId,omitempty" example:"42"`
	// SpaceID は kind=space_all のときだけ入る。
	SpaceID *string `json:"spaceId,omitempty"`
	// Name は kind=group のときだけ入る。
	Name      string    `json:"name" example:"開発チーム"`
	CreatedAt time.Time `json:"createdAt"`
}

func toKbPrincipalResponse(p *domain.Principal) kbPrincipalResponse {
	return kbPrincipalResponse{
		ID:        p.ID,
		Kind:      string(p.Kind),
		UserID:    p.UserID,
		SpaceID:   p.SpaceID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
	}
}

// kbCreateGroupRequest はグループ作成の入力。
type kbCreateGroupRequest struct {
	Name string `json:"name" binding:"required,max=200" example:"開発チーム"`
}

// kbUserIDParam は URL の userId を読む。読めなければ応答を書いて ok=false を返す。
//
// **必ず認可を通したあとで呼ぶこと。** 形式不正（400）と存在しない（404）を撃ち分けるので、
// 認可より先に呼ぶと、権限の無い相手にも「この値は形式としては正しい」が漏れる。
func kbUserIDParam(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return 0, false
	}
	return id, true
}

// AddMember はユーザーをワークスペースのメンバーにする（冪等）。
//
//	@Summary      ナレッジ 基盤 の メンバー 追加
//	@Description  ユーザー を ワークスペース の メンバー に する。 所属 は principals (kind='user') の 行 が 唯一 の 表現 な の で、 この API は その 行 を 作る (既に あれ ば それ を 返す)。 所属 する だけ で は 何 も 見え ない (役割 が 1 つ も 無い) の で、 続け て 権限 付与 の API を 呼ぶ。 呼べる の は ワークスペース の admin だけ で、 権限 が 無い 場合 と 対象 が 存在 し ない 場合 は 同じ 404。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        userId         path      integer true  "ユーザー ID"
//	@Success      200            {object}  kbPrincipalResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/members/{userId} [put]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) AddMember(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	userID, ok := kbUserIDParam(c)
	if !ok {
		return
	}
	// 既知の残り穴: この経路は users.id をそのまま受け取るので、admin から見ると
	// 成功（200）と 404 の差でユーザーの実在が分かる。ワークスペースを作れば誰でも
	// admin になれる以上、ユーザー ID 空間の走査を完全には塞げていない。
	// 塞ぐには「誰を招けるか」を会社などで絞る必要があり、それは権限モデルの外側の
	// 設計判断になるため別途扱う（このチケットの範囲は既存 usecase の配線）。
	principal, err := h.addMember.Execute(c.Request.Context(), usecase.AddWorkspaceMemberInput{
		WorkspaceID: scope.workspaceID,
		UserID:      userID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPrincipalResponse(principal))
}

// RemoveMember はユーザーをワークスペースから外す（冪等）。
//
//	@Summary      ナレッジ 基盤 の メンバー 削除
//	@Description  ユーザー を ワークスペース から 外す。 主体 を 消す の で、 その 人 に 張ら れ て い た 権限 (grant / 例外 / グループ 所属) も 一緒 に 消える (権限 だけ が 残ら ない)。 元 から 非 メンバー なら 何 も せ ず 成功 する (冪等)。 呼べる の は ワークスペース の admin だけ。 ユーザー の admin が 1 人 も 残ら なく なる とき は 409 で 断る。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path  string   true  "ワークスペース の slug"
//	@Param        userId         path  integer  true  "ユーザー ID"
//	@Success      204            "削除 済み"
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      409            {object}  errorResponse  "最後 の admin を 外す 操作"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/members/{userId} [delete]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) RemoveMember(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	userID, ok := kbUserIDParam(c)
	if !ok {
		return
	}
	// メンバーを外すと principal ごと消え、その主体の grant も CASCADE で消える。
	// つまり grant の取り消しと同じく「最後の admin」を消し得るので、同じ検査を通す。
	if !h.requireNotLastWorkspaceAdminByUser(c, scope, userID) {
		return
	}
	if err := h.removeMember.Execute(c.Request.Context(), usecase.RemoveWorkspaceMemberInput{
		WorkspaceID: scope.workspaceID,
		UserID:      userID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// requireNotLastWorkspaceAdminByUser は「最後の admin を消す操作」を断る（ユーザー指定版）。
// 判断の根拠は KnowledgeBaseGrantHandler.requireNotLastWorkspaceAdmin と
// CanRemoveWorkspaceAdminUseCase の doc を参照。
func (h *KnowledgeBaseMemberHandler) requireNotLastWorkspaceAdminByUser(
	c *gin.Context, scope kbRequestScope, userID uint64,
) bool {
	ok, err := h.canRemoveAdmin.Execute(c.Request.Context(), usecase.CanRemoveWorkspaceAdminInput{
		WorkspaceID: scope.workspaceID,
		UserID:      userID,
	})
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

// CreateGroup は権限をまとめて張るためのグループを作る。
//
//	@Summary      ナレッジ 基盤 の グループ 作成
//	@Description  権限 を まとめ て 張る ため の グループ (kind='group' の 主体) を 作る。 名前 は ワークスペース 内 で 一意 (同名 が 2 つ ある と 権限 を 張る 先 を 人 が 選べ ない)。 グループ の 入れ子 は 作れ ない。 呼べる の は ワークスペース の admin だけ。
//	@Tags         knowledge-base
//	@Accept       json
//	@Produce      json
//	@Param        workspaceSlug  path      string                true  "ワークスペース の slug"
//	@Param        body           body      kbCreateGroupRequest  true  "グループ 名"
//	@Success      201            {object}  kbPrincipalResponse
//	@Failure      400            {object}  errorResponse  "バリデーション エラー"
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      409            {object}  errorResponse  "グループ 名 が 使用 済み"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/groups [post]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) CreateGroup(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	limitKnowledgeBaseBody(c)
	var req kbCreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}
	principal, err := h.createGroup.Execute(c.Request.Context(), usecase.CreatePrincipalGroupInput{
		WorkspaceID: scope.workspaceID,
		Name:        req.Name,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, toKbPrincipalResponse(principal))
}

// AddGroupMember はグループにユーザーを加える（冪等）。
//
//	@Summary      ナレッジ 基盤 の グループ メンバー 追加
//	@Description  グループ に ユーザー を 加える。 加える 相手 を 主体 ID で は なく ユーザー ID で 受ける の は、 この 入口 から グループ の 入れ子 を 作ら せ ない ため (DB 側 も 複合 FK で member を kind='user' に 固定 し て いる)。 対象 が ワークスペース の メンバー で なけれ ば 主体 が 無い の で 404。 呼べる の は ワークスペース の admin だけ。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug     path  string   true  "ワークスペース の slug"
//	@Param        groupPrincipalId  path  string   true  "グループ の 主体 ID (UUID)"
//	@Param        userId            path  integer  true  "ユーザー ID"
//	@Success      204               "追加 済み"
//	@Failure      400               {object}  errorResponse  "バリデーション エラー / 主体 の 種類 違い"
//	@Failure      401               {object}  errorResponse  "未 認証"
//	@Failure      404               {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500               {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/groups/{groupPrincipalId}/members/{userId} [put]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) AddGroupMember(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	userID, ok := kbUserIDParam(c)
	if !ok {
		return
	}
	if err := h.addGroupMember.Execute(c.Request.Context(), usecase.AddGroupMemberInput{
		WorkspaceID:      scope.workspaceID,
		GroupPrincipalID: c.Param("groupPrincipalId"),
		MemberUserID:     userID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RemoveGroupMember はグループからユーザーを外す（冪等）。
//
//	@Summary      ナレッジ 基盤 の グループ メンバー 削除
//	@Description  グループ から ユーザー を 外す。 元 から 載っ て い なけれ ば 何 も せ ず 成功 する (冪等)。 グループ 宛て の admin は 「最後 の admin」 の 数 に 入れ て い ない の で、 この 操作 が 409 に なる こと は ない。 呼べる の は ワークスペース の admin だけ。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug     path  string   true  "ワークスペース の slug"
//	@Param        groupPrincipalId  path  string   true  "グループ の 主体 ID (UUID)"
//	@Param        userId            path  integer  true  "ユーザー ID"
//	@Success      204               "削除 済み"
//	@Failure      400               {object}  errorResponse  "バリデーション エラー"
//	@Failure      401               {object}  errorResponse  "未 認証"
//	@Failure      404               {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500               {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/groups/{groupPrincipalId}/members/{userId} [delete]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) RemoveGroupMember(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	if !h.requireWorkspaceAdmin(c, scope) {
		return
	}
	userID, ok := kbUserIDParam(c)
	if !ok {
		return
	}
	if err := h.removeGroupMember.Execute(c.Request.Context(), usecase.RemoveGroupMemberInput{
		WorkspaceID:      scope.workspaceID,
		GroupPrincipalID: c.Param("groupPrincipalId"),
		MemberUserID:     userID,
	}); err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// EnsureSpaceEveryone はスペースの「全員」を表す主体を用意して返す（冪等）。
//
//	@Summary      ナレッジ 基盤 の スペース 全員 主体 の 用意
//	@Description  「既定 で チーム 全員 が 編集 できる」 を grant 1 行 で 表す ため の 主体 (kind='space_all') を 用意 し て 返す。 既に あれ ば それ を 返す (冪等)。 スペース 作成 時 に は 作ら ない 設計 な の で、 権限 を 張る 直前 に この API で ID を 得る。 呼べる の は その スペース の admin (ワークスペース の admin を 含む) だけ。
//	@Tags         knowledge-base
//	@Produce      json
//	@Param        workspaceSlug  path      string  true  "ワークスペース の slug"
//	@Param        spaceId        path      string  true  "スペース ID (UUID)"
//	@Success      200            {object}  kbPrincipalResponse
//	@Failure      401            {object}  errorResponse  "未 認証"
//	@Failure      404            {object}  errorResponse  "権限 が 無い か 対象 が 無い"
//	@Failure      500            {object}  errorResponse  "DB 失敗"
//	@Router       /kb/workspaces/{workspaceSlug}/spaces/{spaceId}/principals/everyone [put]
//	@Security     CookieAuth
func (h *KnowledgeBaseMemberHandler) EnsureSpaceEveryone(c *gin.Context) {
	scope, ok := kbScope(c)
	if !ok {
		return
	}
	spaceID := c.Param("spaceId")
	if !h.requireSpaceAdmin(c, scope, spaceID) {
		return
	}
	principal, err := h.ensureEveryone.Execute(c.Request.Context(), usecase.EnsureSpaceEveryonePrincipalInput{
		WorkspaceID: scope.workspaceID,
		SpaceID:     spaceID,
	})
	if err != nil {
		respondKbPermissionOperationErr(c, err)
		return
	}
	c.JSON(http.StatusOK, toKbPrincipalResponse(principal))
}
