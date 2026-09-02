package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// CourseRepository はコースの永続化を担う。
type CourseRepository interface {
	// ListByWorkspaceID はワークスペース単位のコース一覧を返す。
	ListByWorkspaceID(ctx context.Context, workspaceID string, includeUnpublished bool) ([]domain.Course, error)
	GetByID(ctx context.Context, id uint64) (*domain.Course, error)
	// Create はコースを 1 件作る。
	Create(ctx context.Context, c *domain.Course) error
	// CreateWithOwnerGrant はコースを作り、**同じトランザクションで**作成者に admin の
	// 付与を入れる。
	//
	// 2 つに分けない理由は、途中で落ちたときに「誰も編集できないコース」が残るため。
	// コースを作ることと、作った人がそれを扱えることは 1 つの操作として扱う
	// （共有リンクの発行が主体とリンクを一緒に作るのと同じ）。
	CreateWithOwnerGrant(ctx context.Context, c *domain.Course, ownerPrincipalID string) error
	Update(ctx context.Context, c *domain.Course) error
	Delete(ctx context.Context, id uint64) error
}
