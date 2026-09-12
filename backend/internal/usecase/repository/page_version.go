package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// PageVersionRepository は page_versions テーブルへのアクセスを提供する。
// KnowledgeBaseRepository とは別の fat interface にする（版はページ本文とは別の生存期間・
// 書き込みタイミングを持つテーブルの塊で、これ以上メソッドを足すと肥大しすぎる）。
type PageVersionRepository interface {
	// LockPage は pages 行を SELECT ... FOR UPDATE でロックするだけの操作（CreateVersionIfDue の
	// 手順1と同じクエリ）。呼び出し元は txManager.DoInTx の中でロック→ページの今の内容を読む→
	// CreateVersionIfDue で挿入、の順に呼ぶ。「版を残す」は本文を変えず TouchPageLastEditedBy を
	// 経由しないため、ロックより先に読むと本物の編集が割り込み古い内容のまま版を切る競合があった。
	LockPage(ctx context.Context, workspaceID, pageID string) error

	// CreateVersionIfDue は「版を切るべきか」を判定し、切るべきときだけ 1 件挿入する。
	// pages 行を SELECT ... FOR UPDATE でロックしてから（無ければ repository.ErrPageNotFound。
	// これで本文保存・版を残す・復元の 3 経路を直列化し seq の PK 衝突を防ぐ）直近の版を見て、
	// force || 版が無い || 直近から 10 分超なら次の seq で挿入し、同じトランザクションで
	// 30 日より古い版を削除する。force=true は「版を残す」・復元が使い 10 分規則を無視する。
	// note は保存前に domain.ValidateVersionNote を通した値を渡すこと（ここでは検証しない）。
	CreateVersionIfDue(
		ctx context.Context,
		workspaceID, pageID string,
		doc string,
		authorUserID uint64,
		note *string,
		force bool,
	) (created bool, version *domain.PageVersion, err error)

	// ListVersions はそのページの版一覧を seq 降順で返す。上限 5000 件（30日保持×10分規則の
	// 理論上の最大 4320 に余裕を持たせた defensive な LIMIT）。doc を含まない軽量な行のため
	// ページネーションは今回作らない。
	ListVersions(ctx context.Context, workspaceID, pageID string) ([]domain.PageVersion, error)

	// GetVersion は 1 件の版を返す。無ければ domain.ErrPageVersionNotFound。
	GetVersion(ctx context.Context, workspaceID, pageID string, seq int64) (*domain.PageVersion, error)

	// GetLatestVersion はそのページの直近の版を返す。GetVersion と違い、版が 1 つも無いことは
	// エラーではなく正常として扱う（版の無いページへの提案は BaseSeq を nil にするため
	// CreateSuggestionUseCase が使う）。版が無ければ (nil, nil)。
	GetLatestVersion(ctx context.Context, workspaceID, pageID string) (*domain.PageVersion, error)
}
