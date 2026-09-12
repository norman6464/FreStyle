package kb

import (
	"context"
	"log/slog"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// AddPageFavoriteUseCase はページをお気に入りに付ける。
//
// 見えないページを付けさせない判定は、他の 1 ページ操作（SetIcon 等）と同じく handler 側の
// requirePagePermission（CapabilityView）が担う。ここでは付けた行が今回新しくできたか
// （201 vs 204 の判定に使う）だけを返す。
type AddPageFavoriteUseCase struct {
	repo repository.PageFavoriteRepository
}

func NewAddPageFavoriteUseCase(r repository.PageFavoriteRepository) *AddPageFavoriteUseCase {
	return &AddPageFavoriteUseCase{repo: r}
}

func (u *AddPageFavoriteUseCase) Execute(ctx context.Context, workspaceID, pageID string, userID uint64) (created bool, err error) {
	return u.repo.Add(ctx, workspaceID, pageID, userID)
}

// IsPageFavoriteUseCase は resolve 応答の isFavorite（★ の初期状態）に答える。
type IsPageFavoriteUseCase struct {
	repo repository.PageFavoriteRepository
}

func NewIsPageFavoriteUseCase(r repository.PageFavoriteRepository) *IsPageFavoriteUseCase {
	return &IsPageFavoriteUseCase{repo: r}
}

func (u *IsPageFavoriteUseCase) Execute(ctx context.Context, pageID string, userID uint64) (bool, error) {
	return u.repo.IsFavorite(ctx, pageID, userID)
}

// RemovePageFavoriteUseCase は外す。冪等（付いていなくても成功扱い）。
type RemovePageFavoriteUseCase struct {
	repo repository.PageFavoriteRepository
}

func NewRemovePageFavoriteUseCase(r repository.PageFavoriteRepository) *RemovePageFavoriteUseCase {
	return &RemovePageFavoriteUseCase{repo: r}
}

func (u *RemovePageFavoriteUseCase) Execute(ctx context.Context, pageID string, userID uint64) error {
	return u.repo.Remove(ctx, pageID, userID)
}

// ListPageFavoritesUseCase はそのワークスペース内の自分のお気に入りを、可視判定でふるってから
// 付けた順の新しい順に返す。お気に入りにした後で権限を失ったページは応答から落とす
// （行は消さない — ListMyRecentPagesUseCase と同じ考え方）。
type ListPageFavoritesUseCase struct {
	repo  repository.PageFavoriteRepository
	check *CheckPagePermissionUseCase
}

func NewListPageFavoritesUseCase(r repository.PageFavoriteRepository, check *CheckPagePermissionUseCase) *ListPageFavoritesUseCase {
	return &ListPageFavoritesUseCase{repo: r, check: check}
}

func (u *ListPageFavoritesUseCase) Execute(ctx context.Context, workspaceID string, userID uint64) ([]domain.PageFavorite, error) {
	candidates, err := u.repo.ListFavorites(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PageFavorite, 0, len(candidates))
	for _, c := range candidates {
		perm, err := u.check.Execute(ctx, CheckPagePermissionInput{
			WorkspaceID: workspaceID, PageID: c.PageID, UserID: userID,
		})
		if err != nil {
			slog.WarnContext(ctx, "kb: favorite permission check failed", "pageID", c.PageID, "err", err)
			continue
		}
		if !perm.CanView {
			continue
		}
		out = append(out, c)
	}
	return out, nil
}
