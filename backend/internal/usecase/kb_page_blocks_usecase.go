package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ReplacePageBlocksUseCase はページ本文（ProseMirror doc）をブロック行に分解して
// 全入れ替えし、snapshot を焼き直す。差分更新は将来の最適化で、まず全消し全入れで正しさを取る。
//
// 保存する snapshot は入力 doc そのものではなく、分解した木から組み立て直した正規形。
// 入力に行スキーマへ写せない情報（未知フィールド等）が混ざっていても
// 「snapshot は必ず blocks から再生成できる」という不変条件が崩れないようにするため。
type ReplacePageBlocksUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewReplacePageBlocksUseCase(r repository.KnowledgeBaseRepository) *ReplacePageBlocksUseCase {
	return &ReplacePageBlocksUseCase{repo: r}
}

type ReplacePageBlocksInput struct {
	WorkspaceID string
	PageID      string
	// Doc は ProseMirror ドキュメント（tiptap の getJSON() 相当の JSON 文字列）。
	Doc string
}

func (u *ReplacePageBlocksUseCase) Execute(ctx context.Context, in ReplacePageBlocksInput) (*domain.PageSnapshot, error) {
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if page.ArchivedAt != nil {
		return nil, ErrPageArchived
	}
	tree, err := parsePageDoc(in.Doc)
	if err != nil {
		return nil, err
	}
	rows, err := flattenPageDoc(tree)
	if err != nil {
		return nil, err
	}
	normalized, err := renderPageDoc(tree)
	if err != nil {
		return nil, err
	}
	if err := u.repo.ReplacePageBlocks(ctx, in.WorkspaceID, in.PageID, rows, normalized); err != nil {
		return nil, err
	}
	return u.repo.GetPageSnapshot(ctx, in.WorkspaceID, in.PageID)
}
