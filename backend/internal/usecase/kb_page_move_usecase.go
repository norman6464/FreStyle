package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// MovePageUseCase はページ（とその子孫）を別の親・別のスペースへ移す。
// 循環検出 → 移動先末尾の position 採番 → pages / page_paths / 子孫 space_id の
// 一括更新（repository 内の 1 トランザクション）の順で行う。
type MovePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewMovePageUseCase(r repository.KnowledgeBaseRepository) *MovePageUseCase {
	return &MovePageUseCase{repo: r}
}

type MovePageInput struct {
	WorkspaceID string
	PageID      string
	// NewParentID が nil ならスペース直下（ルート）へ移す。
	NewParentID *string
	// NewSpaceID は NewParentID が nil のときの移動先スペース。空なら現在のスペースの
	// ルートへ移す。NewParentID があるときは親の所属スペースが移動先になる
	// （指定があれば親と一致することを検証する）。
	NewSpaceID string
}

func (u *MovePageUseCase) Execute(ctx context.Context, in MovePageInput) (*domain.Page, error) {
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if page.ArchivedAt != nil {
		return nil, ErrPageArchived
	}

	var targetSpaceID string
	if in.NewParentID != nil {
		if *in.NewParentID == in.PageID {
			return nil, ErrPageCycle
		}
		parent, err := u.repo.FindPage(ctx, in.WorkspaceID, *in.NewParentID)
		if err != nil {
			return nil, err
		}
		if parent.ArchivedAt != nil {
			return nil, ErrPageParentArchived
		}
		if in.NewSpaceID != "" && in.NewSpaceID != parent.SpaceID {
			return nil, ErrPageParentSpaceMismatch
		}
		targetSpaceID = parent.SpaceID
		// 新しい親が自分の子孫だと木が根から切り離されて循環する。closure（page_paths）の
		// 1 回の存在確認で検出する（depth=0 の自分自身も含まれる）。
		isDesc, err := u.repo.HasDescendant(ctx, in.WorkspaceID, in.PageID, *in.NewParentID)
		if err != nil {
			return nil, err
		}
		if isDesc {
			return nil, ErrPageCycle
		}
	} else {
		targetSpaceID = in.NewSpaceID
		if targetSpaceID == "" {
			targetSpaceID = page.SpaceID
		}
		if targetSpaceID != page.SpaceID {
			// 別スペースのルートへ移す場合のみ、移動先スペースの実在を確認する
			// （同一スペースなら page 自身の存在が実在の証明になっている）。
			if _, err := u.repo.FindSpace(ctx, in.WorkspaceID, targetSpaceID); err != nil {
				return nil, err
			}
		}
	}

	last, err := u.repo.LastActiveSiblingPosition(ctx, in.WorkspaceID, targetSpaceID, in.NewParentID)
	if err != nil {
		return nil, err
	}
	pos, err := fracindex.Between(last, "")
	if err != nil {
		return nil, err
	}
	if err := u.repo.MovePage(ctx, in.WorkspaceID, in.PageID, in.NewParentID, targetSpaceID, pos); err != nil {
		return nil, err
	}
	return u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
}

// ArchivePageUseCase はページとその子孫をまとめてアーカイブする（ツリーから隠す）。
// 既にアーカイブ済みなら何もしない（冪等）。
type ArchivePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewArchivePageUseCase(r repository.KnowledgeBaseRepository) *ArchivePageUseCase {
	return &ArchivePageUseCase{repo: r}
}

type ArchivePageInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ArchivePageUseCase) Execute(ctx context.Context, in ArchivePageInput) error {
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return err
	}
	if page.ArchivedAt != nil {
		return nil // 冪等: 二重アーカイブで子孫の archived_at を上書きしない
	}
	return u.repo.ArchivePageSubtree(ctx, in.WorkspaceID, in.PageID)
}

// UnarchivePageUseCase はアーカイブしたページを（同時にアーカイブされた子孫ごと）現役へ戻す。
// 部分 UNIQUE は現役の並びだけを守るため、アーカイブ中に同じ position の兄弟ができていたら
// 末尾へ再採番してから戻す。親がまだアーカイブ中の場合は戻せない（ツリーに現れない
// 「迷子ページ」を作らないため、先に親を戻す運用）。
type UnarchivePageUseCase struct {
	repo repository.KnowledgeBaseRepository
}

func NewUnarchivePageUseCase(r repository.KnowledgeBaseRepository) *UnarchivePageUseCase {
	return &UnarchivePageUseCase{repo: r}
}

type UnarchivePageInput struct {
	WorkspaceID string
	PageID      string
}

func (u *UnarchivePageUseCase) Execute(ctx context.Context, in UnarchivePageInput) (*domain.Page, error) {
	page, err := u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if page.ArchivedAt == nil {
		return page, nil // 冪等
	}
	if page.ParentID != nil {
		parent, err := u.repo.FindPage(ctx, in.WorkspaceID, *page.ParentID)
		if err != nil {
			return nil, err
		}
		if parent.ArchivedAt != nil {
			return nil, ErrPageParentArchived
		}
	}
	var newPos *string
	conflicted, err := u.repo.HasActiveSiblingPosition(ctx, in.WorkspaceID, page.SpaceID, page.ParentID, page.Position, page.ID)
	if err != nil {
		return nil, err
	}
	if conflicted {
		last, err := u.repo.LastActiveSiblingPosition(ctx, in.WorkspaceID, page.SpaceID, page.ParentID)
		if err != nil {
			return nil, err
		}
		pos, err := fracindex.Between(last, "")
		if err != nil {
			return nil, err
		}
		newPos = &pos
	}
	if err := u.repo.UnarchivePageSubtree(ctx, in.WorkspaceID, in.PageID, *page.ArchivedAt, newPos); err != nil {
		return nil, err
	}
	return u.repo.FindPage(ctx, in.WorkspaceID, in.PageID)
}
