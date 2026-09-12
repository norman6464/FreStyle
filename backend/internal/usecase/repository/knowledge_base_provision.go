package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// WorkspaceProvisionInput はワークスペースの作成に渡す値。ID を持たないのは
// 採番（UUIDv7）が repository の責務のため。
type WorkspaceProvisionInput struct {
	// Slug は URL に出る識別子（グローバルに一意）。
	Slug string
	Name string
	// OwnerUserID は作成者。作成と同じトランザクションで主体（kind='user'）になり、admin の grant を受け取る。
	OwnerUserID uint64
	// PersonalOwnerUserID はサインアップで自動作成する個人ワークスペースのときだけ設定する
	// （通常は nil）。workspaces.personal_owner_user_id に書かれ、1 人 1 つを
	// uq_workspaces_personal_owner が守る。
	PersonalOwnerUserID *uint64
}

// WorkspaceProvisioner はワークスペースを「入れる人ごと」作る単一責務の port。
// KnowledgeBaseRepository でも KnowledgeBasePermissionRepository でもなく別にしているのは、
// この操作が意図的に両方の境界をまたぐため（どちらかの fat interface に足すと、その実装が
// 相手側のテーブルを書くことになる）。
//
// 3 つの書き込み（workspaces / principals / workspace_grants）は 1 トランザクションで行う。
// 分けると、主体と grant が入らないまま誰もメンバーでないワークスペースができ、
// middleware の所属確認により作成者にも 404 にしか見えず、slug だけを占有し続ける
// （消す口も無い）。
type WorkspaceProvisioner interface {
	// ProvisionWorkspace はワークスペースを作り、作成者を admin のメンバーにして返す。
	// slug が使用済みなら ErrWorkspaceSlugTaken。
	ProvisionWorkspace(ctx context.Context, in WorkspaceProvisionInput) (*domain.Workspace, error)
	// ProvisionPrivateSpace はプライベートスペースを作り、作成者へ space_grant(admin) を張って返す。
	// key が使用済みなら ErrSpaceKeyTaken、作成者が非メンバーなら ErrPrincipalNotFound。
	// spaces / space_grants の 2 書き込みは 1 トランザクション（private はワークスペース既定の
	// grant が届かないため、分けると作った本人にも見えないスペースが key だけ占有して残る）。
	ProvisionPrivateSpace(ctx context.Context, in PrivateSpaceProvisionInput) (*domain.Space, error)
}

// PrivateSpaceProvisionInput はプライベートスペースの作成に渡す値。
// ID を持たないのは採番（UUIDv7）が repository の責務のため。
type PrivateSpaceProvisionInput struct {
	WorkspaceID string
	// Key はワークスペース内で一意な短い識別子。呼び出し側（usecase）が検証・採番済み。
	Key  string
	Name string
	// CreatorUserID は作成者。既にワークスペースのメンバー（principals の行がある）であること。
	CreatorUserID uint64
}
