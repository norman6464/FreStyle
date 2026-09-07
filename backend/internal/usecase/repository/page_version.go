package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// PageVersionRepository は page_versions テーブルへのアクセスを提供する。
//
// KnowledgeBaseRepository とは別の fat interface にする（comment.CommentRepository と
// 同じ役割分担 — 版はページ本文とは別の生存期間・別の書き込みタイミングを持つテーブルの塊で、
// KnowledgeBaseRepository にこれ以上メソッドを足すと 1 interface が肥大しすぎる）。
type PageVersionRepository interface {
	// CreateVersionIfDue は「版を切るべきか」を判定し、切るべきときだけ 1 件挿入する。
	//
	// 判定は次の順で行う（実装は persistence 層。usecase からは判定結果しか見えない）:
	//  1. pages 行を SELECT ... FOR UPDATE でロックする（対象ページが無ければ
	//     repository.ErrPageNotFound。これが「本文保存経路」「版を残す」「復元」の 3 経路を
	//     直列化し、seq の PK 衝突を原理的に起こさせない — ロック順は常に pages が先）。
	//  2. そのページの直近の版（seq 降順 1 件）を取る。
	//  3. shouldCut := force || 直近の版が無い || now - 直近の版.CreatedAt > 10分。
	//     shouldCut が false なら created=false, version=nil, err=nil を返す（何もしない）。
	//  4. shouldCut なら seq = 直近 + 1（無ければ 1）で 1 件挿入する。
	//  5. 挿入と同じトランザクションで、そのページの 30 日より古い版を削除する
	//     （掃除。今挿入した行は created_at が新しいので対象にならない）。
	//
	// force=true は「版を残す」操作・復元が使う — 10 分規則を無視して必ず切る。
	// note は保存前に domain.ValidateVersionNote を通した後の値を渡すこと（ここでは検証しない）。
	CreateVersionIfDue(
		ctx context.Context,
		workspaceID, pageID string,
		doc string,
		authorUserID uint64,
		note *string,
		force bool,
	) (created bool, version *domain.PageVersion, err error)

	// ListVersions はそのページの版一覧を seq 降順で返す。
	//
	// 上限 500 件（defensive な LIMIT。ページネーションは今回作らない）。
	// ページの版が 500 を超える場合、501 件目以降は一覧に出ない。
	ListVersions(ctx context.Context, workspaceID, pageID string) ([]domain.PageVersion, error)

	// GetVersion は 1 件の版を返す。無ければ domain.ErrPageVersionNotFound。
	GetVersion(ctx context.Context, workspaceID, pageID string, seq int64) (*domain.PageVersion, error)
}
