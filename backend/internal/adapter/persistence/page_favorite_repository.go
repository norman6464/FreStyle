package persistence

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// pageFavoriteRepository は [repository.PageFavoriteRepository] の実装。page_favorites は
// KnowledgeBaseRepository と同じくスキーマの正本が schema.hcl で GORM を通さない方針のため、
// クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type pageFavoriteRepository struct {
	baseRepository
}

// NewPageFavoriteRepository はお気に入りの repository を組み立てる。
func NewPageFavoriteRepository(db *sql.DB) repository.PageFavoriteRepository {
	return &pageFavoriteRepository{baseRepository{db: db}}
}

func (r *pageFavoriteRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func (r *pageFavoriteRepository) Add(ctx context.Context, workspaceID, pageID string, userID uint64) (bool, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return false, repository.ErrPageNotFound
	}
	uid, ok3 := toInt64ID(userID)
	if !ok3 {
		return false, outOfRangeIDError("user_id", userID)
	}
	n, err := r.queries(ctx).AddPageFavorite(ctx, sqlcgen.AddPageFavoriteParams{
		UserID:      uid,
		WorkspaceID: wsID,
		PageID:      pgID,
	})
	if isForeignKeyViolation(err) {
		// RecordView と同じ理由（page_view_repository.go 参照）: workspaceID と pageID が
		// 噛み合わない、または pageID が実在しない場合は fk_page_favorites_page が弾く。
		return false, repository.ErrPageNotFound
	}
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *pageFavoriteRepository) Remove(ctx context.Context, pageID string, userID uint64) error {
	pgID, ok := kbParseID(pageID)
	if !ok {
		// 形式の壊れた ID は「もともと付いていない」と同じ扱いにする（外すのは冪等）。
		return nil
	}
	uid, ok2 := toInt64ID(userID)
	if !ok2 {
		return outOfRangeIDError("user_id", userID)
	}
	return r.queries(ctx).RemovePageFavorite(ctx, sqlcgen.RemovePageFavoriteParams{
		UserID: uid,
		PageID: pgID,
	})
}

func (r *pageFavoriteRepository) IsFavorite(ctx context.Context, pageID string, userID uint64) (bool, error) {
	pgID, ok := kbParseID(pageID)
	if !ok {
		return false, nil
	}
	uid, ok2 := toInt64ID(userID)
	if !ok2 {
		return false, outOfRangeIDError("user_id", userID)
	}
	return r.queries(ctx).IsPageFavorite(ctx, sqlcgen.IsPageFavoriteParams{
		UserID: uid,
		PageID: pgID,
	})
}

func (r *pageFavoriteRepository) ListFavorites(ctx context.Context, workspaceID string, userID uint64) ([]domain.PageFavorite, error) {
	wsID, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrPageNotFound
	}
	uid, ok2 := toInt64ID(userID)
	if !ok2 {
		return nil, outOfRangeIDError("user_id", userID)
	}
	rows, err := r.queries(ctx).ListPageFavorites(ctx, sqlcgen.ListPageFavoritesParams{
		UserID:      uid,
		WorkspaceID: wsID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PageFavorite, 0, len(rows))
	for _, row := range rows {
		pf := domain.PageFavorite{
			PageID:    row.PageID.String(),
			Title:     row.Title,
			SpaceID:   row.SpaceID.String(),
			SpaceName: row.SpaceName,
			CreatedAt: row.CreatedAt,
		}
		// icon は飾りであって、壊れていても一覧の読み出しを止める理由にはならない
		// （toDomainPage と同じ扱い）。
		if row.Icon != nil {
			var icon domain.PageIcon
			if err := json.Unmarshal(*row.Icon, &icon); err == nil {
				pf.Icon = &icon
			}
		}
		out = append(out, pf)
	}
	return out, nil
}
