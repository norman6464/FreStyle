package repository

import (
	"context"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// PageSuggestionRepository は page_suggestions テーブルへのアクセスを提供する。
//
// KnowledgeBaseRepository とは別の fat interface にする（PageTemplateRepository と同じ役割
// 分担 — 提案はページ本文とは別の生存期間を持つ独立した資産の塊で、KnowledgeBaseRepository に
// これ以上メソッドを足すと 1 interface が肥大しすぎる）。
type PageSuggestionRepository interface {
	// Create は提案を 1 件作る。ID は repository 側で採番する（page_templates と同じ流儀）。
	Create(ctx context.Context, s *domain.PageSuggestion) error
	// ListOpen はそのページの open な提案を created_at 昇順で最大 limit 件返す。limit は
	// SQL の LIMIT にそのまま渡す（呼び出し元の usecase が上限を挟んでから渡すため、ここでは
	// 検証しない）。大量の open 提案が積まれても、doc 込みの全件を一度にメモリへ載せない。
	ListOpen(ctx context.Context, workspaceID, pageID string, limit int) ([]domain.PageSuggestion, error)
	// CountOpen はそのページの open な提案の総数。
	CountOpen(ctx context.Context, workspaceID, pageID string) (int, error)
	// CountOpenByAuthor はそのページ・その投稿者本人の open な提案数
	// （投稿者 1 人あたりの上限判定に使う）。
	CountOpenByAuthor(ctx context.Context, workspaceID, pageID string, authorUserID uint64) (int, error)
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
