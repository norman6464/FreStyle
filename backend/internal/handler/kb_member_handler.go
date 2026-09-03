package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// KnowledgeBaseMemberHandler はノートの主体（principals）の出し入れを受ける。
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
// workspaceId は載せない（URL の slug で決まる）。id は載せる — grant を
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
