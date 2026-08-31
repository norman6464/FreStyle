package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// WorkspaceActivationReader は所属ワークスペースが停止されていないかを見るための最小の読み取り口。
//
// middleware が全 API の入口で使う。KnowledgeBaseRepository をそのまま渡すと、入口の
// 認可検査がノートの読み書きにまで手が届いてしまうので、必要な 1 メソッドだけ切り出す。
type WorkspaceActivationReader interface {
	// FindWorkspaceByID はワークスペースを 1 件引く。無ければ ErrWorkspaceNotFound。
	FindWorkspaceByID(ctx context.Context, workspaceID string) (*domain.Workspace, error)
}
