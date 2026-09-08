package repository

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// PageSuggestionRepository は page_suggestions テーブルへのアクセスを提供する。
//
// KnowledgeBaseRepository とは別の fat interface にする（PageTemplateRepository と同じ役割
// 分担 — 提案はページ本文とは別の生存期間を持つ独立した資産の塊で、KnowledgeBaseRepository に
// これ以上メソッドを足すと 1 interface が肥大しすぎる）。
type PageSuggestionRepository interface {
	// Create は提案を 1 件作る。ID は repository 側で採番する（page_templates と同じ流儀）。
	Create(ctx context.Context, s *domain.PageSuggestion) error
	// ListOpen はそのページの open な提案を created_at 昇順で返す。
	ListOpen(ctx context.Context, workspaceID, pageID string) ([]domain.PageSuggestion, error)
	// Get は提案を 1 件取得する（accept/reject の直前に呼ぶ）。
	Get(ctx context.Context, workspaceID, pageID, suggestionID string) (*domain.PageSuggestion, error)
	// Resolve は open の提案だけを指定の status へ条件付き UPDATE する（WHERE status='open'）。
	// 他の誰かが既に解決済みなら domain.ErrPageSuggestionAlreadyResolved、行自体が無ければ
	// domain.ErrPageSuggestionNotFound を返す。戻り値の BaseSeq は必ず nil になる
	// （解決と同時に参照を切るため。queries/page_suggestion.sql の ResolvePageSuggestion 参照）。
	Resolve(
		ctx context.Context, workspaceID, pageID, suggestionID string,
		status domain.PageSuggestionStatus, resolverUserID uint64, resolvedAt time.Time,
	) (*domain.PageSuggestion, error)
}
