package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// PageViewRepository は page_views テーブル（人 × ページの「最後に見た日時」を 1 行だけ持つ）
// へのアクセスを提供する。KnowledgeBaseRepository とは別の fat interface にする
// （PageVersionRepository と同じ役割分担 — 閲覧記録はページ本文とは別の生存期間・別の
// 書き込みタイミングを持つテーブルの塊）。
type PageViewRepository interface {
	// RecordView は (userID, pageID) の行を upsert する（無ければ作り、あれば viewed_at を
	// now() へ進める）。pageID が実在しない、または workspaceID と噛み合わなければ
	// repository.ErrPageNotFound。
	RecordView(ctx context.Context, workspaceID, pageID string, userID uint64) error

	// CountViews はそのページを見たことのある人数（= 行数）を返す。upsert 型の表なので
	// 延べ回数ではない。
	CountViews(ctx context.Context, pageID string) (int, error)

	// ListRecentPageViewCandidates は userID の「最近見たページ」候補を viewed_at の新しい順に
	// 返す（ワークスペース横断。件数は実装内部で頭打ちにする）。可視判定はここでは行わない
	// — 呼び出し元の usecase が CheckPagePermissionUseCase を通してからふるう。
	ListRecentPageViewCandidates(ctx context.Context, userID uint64) ([]domain.RecentPage, error)
}
