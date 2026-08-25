package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrPagePermissionDenied はページに対する操作が実効権限で許されていないときに返す。
// handler はこれを 403（あるいは存在自体を隠すなら 404）にマップする。
var ErrPagePermissionDenied = errors.New("permission denied for this page")

// CheckPagePermissionUseCase は「このユーザーはこのページを閲覧 / 編集できるか」に答える。
// ナレッジ基盤の認可はすべてここを通す（呼び出し側に判定規則を写経させない）。
//
// 段 1-b の各 usecase（GetPageUseCase / RenamePageUseCase / ReplacePageBlocksUseCase …）への
// 組み込みは handler の段で行う。組み込み方は次のとおり:
//
//	perm, err := check.Execute(ctx, usecase.CheckPagePermissionInput{
//	    WorkspaceID: workspaceID, PageID: pageID, UserID: currentUserID,
//	})
//	if err != nil {
//	    return err // ページが無い場合は repository.ErrPageNotFound がそのまま来る
//	}
//	if !perm.CanView { // 書き込み系なら !perm.CanEdit
//	    return usecase.ErrPagePermissionDenied
//	}
//
// ツリー取得のように複数ページを扱う経路では、ページごとにこれを呼ばず
// ListViewablePagesUseCase を使う（1 ページ 1 往復にしないため）。
type CheckPagePermissionUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewCheckPagePermissionUseCase(r repository.KnowledgeBasePermissionRepository) *CheckPagePermissionUseCase {
	return &CheckPagePermissionUseCase{repo: r}
}

type CheckPagePermissionInput struct {
	WorkspaceID string
	PageID      string
	UserID      uint64
}

func (u *CheckPagePermissionUseCase) Execute(ctx context.Context, in CheckPagePermissionInput) (*domain.PagePermission, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.PageID == "" {
		return nil, errors.New("pageID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	facts, err := u.repo.PagePermissionFactsForUser(ctx, in.WorkspaceID, in.PageID, in.UserID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolvePagePermission(*facts)
	return &perm, nil
}

// IsWorkspaceMemberUseCase は「このユーザーはこのワークスペースのメンバーか」に答える。
// 所属は principals（kind='user'）の行の有無がすべてで、専用のメンバーシップ表は無い。
type IsWorkspaceMemberUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewIsWorkspaceMemberUseCase(r repository.KnowledgeBasePermissionRepository) *IsWorkspaceMemberUseCase {
	return &IsWorkspaceMemberUseCase{repo: r}
}

type IsWorkspaceMemberInput struct {
	WorkspaceID string
	UserID      uint64
}

func (u *IsWorkspaceMemberUseCase) Execute(ctx context.Context, in IsWorkspaceMemberInput) (bool, error) {
	if in.WorkspaceID == "" {
		return false, errors.New("workspaceID is required")
	}
	if in.UserID == 0 {
		return false, errors.New("userID is required")
	}
	return u.repo.IsWorkspaceMember(ctx, in.WorkspaceID, in.UserID)
}

// ListViewablePagesUseCase はスペース配下の現役ページのうち、そのユーザーが閲覧できるものを返す。
// ツリー取得の土台。ページ数によらず問い合わせは 1 回で、判定は
// CheckPagePermissionUseCase と同じ domain.ResolvePagePermission を通る。
type ListViewablePagesUseCase struct {
	repo repository.KnowledgeBasePermissionRepository
}

func NewListViewablePagesUseCase(r repository.KnowledgeBasePermissionRepository) *ListViewablePagesUseCase {
	return &ListViewablePagesUseCase{repo: r}
}

type ListViewablePagesInput struct {
	WorkspaceID string
	SpaceID     string
	UserID      uint64
}

func (u *ListViewablePagesUseCase) Execute(ctx context.Context, in ListViewablePagesInput) ([]domain.Page, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	rows, err := u.repo.ListSpacePageViewFacts(ctx, in.WorkspaceID, in.SpaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		if domain.ResolvePagePermission(row.Facts).CanView {
			pages = append(pages, row.Page)
		}
	}
	return pages, nil
}
