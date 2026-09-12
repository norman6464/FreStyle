package persistence

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/norman6464/frestyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// pageViewRepository は [repository.PageViewRepository] の実装。page_views は
// KnowledgeBaseRepository と同じくスキーマの正本が schema.hcl で GORM を通さない方針のため、
// クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type pageViewRepository struct {
	baseRepository
}

// NewPageViewRepository は閲覧記録の repository を組み立てる。
func NewPageViewRepository(db *sql.DB) repository.PageViewRepository {
	return &pageViewRepository{baseRepository{db: db}}
}

func (r *pageViewRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

func (r *pageViewRepository) RecordView(ctx context.Context, workspaceID, pageID string, userID uint64) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	uid, ok3 := toInt64ID(userID)
	if !ok3 {
		return outOfRangeIDError("user_id", userID)
	}
	err := r.queries(ctx).UpsertPageView(ctx, sqlcgen.UpsertPageViewParams{
		UserID:      uid,
		WorkspaceID: wsID,
		PageID:      pgID,
	})
	if isForeignKeyViolation(err) {
		// workspaceID と pageID が噛み合わない（別ワークスペースのページ）、または
		// pageID が実在しない場合は fk_page_views_page が弾く。呼び出し元にとっては
		// 「そのページが見つからない」と同じ意味なので、ErrPageNotFound に翻訳する。
		return repository.ErrPageNotFound
	}
	return err
}

func (r *pageViewRepository) CountViews(ctx context.Context, pageID string) (int, error) {
	pgID, ok := kbParseID(pageID)
	if !ok {
		// 実在しない/形式の壊れた ID は「0 件」として返す。閲覧数は付随情報であり、
		// これのために呼び出し元にエラーを伝播させて本文の表示自体を止める理由が無い。
		return 0, nil
	}
	n, err := r.queries(ctx).CountPageViews(ctx, pgID)
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

func (r *pageViewRepository) ListRecentPageViewCandidates(ctx context.Context, userID uint64) ([]domain.RecentPage, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, outOfRangeIDError("user_id", userID)
	}
	rows, err := r.queries(ctx).ListRecentPageViewCandidates(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.RecentPage, 0, len(rows))
	for _, row := range rows {
		rp := domain.RecentPage{
			PageID:        row.PageID.String(),
			WorkspaceID:   row.WorkspaceID.String(),
			WorkspaceSlug: row.WorkspaceSlug,
			Title:         row.Title,
			SpaceID:       row.SpaceID.String(),
			SpaceName:     row.SpaceName,
			ViewedAt:      row.ViewedAt,
		}
		// icon は飾り（見た目）であって、壊れていても一覧の読み出しを止める理由には
		// ならない（toDomainPage と同じ扱い）。json.Unmarshal に失敗したら nil に倒す。
		if row.Icon != nil {
			var icon domain.PageIcon
			if err := json.Unmarshal(*row.Icon, &icon); err == nil {
				rp.Icon = &icon
			}
		}
		out = append(out, rp)
	}
	return out, nil
}
