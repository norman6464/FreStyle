package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// PageFavoriteRepository は page_favorites テーブルへのアクセスを提供する。
// KnowledgeBaseRepository とは別の fat interface にする（PageViewRepository と同じ役割分担）。
type PageFavoriteRepository interface {
	// Add はお気に入りに付ける（冪等）。created は今回新しく付いたか（既に付いていれば false）。
	// pageID が実在しない、または workspaceID と噛み合わなければ repository.ErrPageNotFound。
	Add(ctx context.Context, workspaceID, pageID string, userID uint64) (created bool, err error)

	// Remove は外す（冪等。行の有無に関わらずエラーにしない）。
	Remove(ctx context.Context, pageID string, userID uint64) error

	// IsFavorite は resolve 応答の isFavorite（★ の初期状態）に使う。
	IsFavorite(ctx context.Context, pageID string, userID uint64) (bool, error)

	// ListFavorites はそのワークスペース内の自分のお気に入りを付けた順の新しい順に返す。
	// 可視判定はここでは行わない — 呼び出し元の usecase が CheckPagePermissionUseCase を
	// 通してからふるう。
	ListFavorites(ctx context.Context, workspaceID string, userID uint64) ([]domain.PageFavorite, error)
}
