package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// WorkspaceProvisionInput はワークスペースの作成に渡す値。
//
// ID を持たないのは採番（UUIDv7）が repository の責務のため。
type WorkspaceProvisionInput struct {
	// Slug は URL に出る識別子（グローバルに一意）。
	Slug string
	// Name は表示名。
	Name string
	// OwnerUserID は作成者。作成と同じトランザクションで主体（kind='user'）になり、
	// admin の grant を受け取る。
	OwnerUserID uint64
}

// WorkspaceProvisioner はワークスペースを「入れる人ごと」作る単一責務の port。
//
// KnowledgeBaseRepository（workspaces を持つ）でも KnowledgeBasePermissionRepository
// （principals / workspace_grants を持つ）でもなく別の port にしているのは、
// この操作が意図的に両方の境界をまたぐため。どちらか一方の fat interface に足すと、
// その実装が相手側のテーブルを書くことになり、境界の意味が薄れる。
// §2.6 が認める「単一責務の port は -er 命名で切り出してよい」に当たる。
//
// 3 つの書き込み（workspaces / principals / workspace_grants）は 1 トランザクションで行う。
// 分けてはいけない: ワークスペースの行だけが入って主体と grant が入らないと、
// 作成者を含めて誰もメンバーでないワークスペースができる。ナレッジ基盤の全経路は
// middleware が所属を確かめてから通すので、そのワークスペースは作成者にも
// 404 にしか見えず、誰も入れないまま slug だけを占有し続ける（消す口も無い）。
type WorkspaceProvisioner interface {
	// ProvisionWorkspace はワークスペースを作り、作成者を admin のメンバーにして返す。
	// slug が使用済みなら ErrWorkspaceSlugTaken。
	ProvisionWorkspace(ctx context.Context, in WorkspaceProvisionInput) (*domain.Workspace, error)
}
