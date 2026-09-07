package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// pageVersionMinInterval は、直近の版からこの時間未満しか経っていなければ新しい版を
// 間引く（作らない）しきい値。CreateVersionIfDue の doc コメント・
// backend タスク設計（FRESTYLE-433 段 3）の「10 分規則」参照。
const pageVersionMinInterval = 10 * time.Minute

// pageVersionRetention は版を残す期間。これより古い版は、新しい版を作るのと同じ
// トランザクションで掃除する（「30 日保持の掃除」）。
const pageVersionRetention = 30 * 24 * time.Hour

// pageVersionRepository は [repository.PageVersionRepository] の実装。page_versions は
// KnowledgeBaseRepository と同じくスキーマの正本が schema.hcl で GORM を通さない方針のため、
// クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
type pageVersionRepository struct {
	baseRepository
}

// NewPageVersionRepository は版の repository を組み立てる。
func NewPageVersionRepository(db *sql.DB) repository.PageVersionRepository {
	return &pageVersionRepository{baseRepository{db: db}}
}

// queries は ctx に乗っているトランザクション（あれば）に束縛した sqlc の Queries を作る。
func (r *pageVersionRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

// runInTx は knowledgeBaseRepository.runInTx と同じパターン。ctx に外側の
// TxManager.DoInTx が開いたトランザクションが既にあればそれへ相乗りし（本文保存経路・復元は
// ReplacePageBlocksUseCase の DoInTx の中からここへ来る）、無ければ自前で開始する
// （「版を残す」は versionRepo 単独の呼び出しで、そちらは自前でトランザクションを持つ）。
// CreateVersionIfDue はロック → 直近の版の取得 → 判定 → 挿入 → 掃除を 1 つの原子的な単位として
// 扱う必要があるため、この repository でも同じパターンを用意する。
func (r *pageVersionRepository) runInTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
	if tx, ok := getTx(ctx); ok {
		return fn(sqlcgen.New(tx))
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op
	if err := fn(sqlcgen.New(tx)); err != nil {
		return err
	}
	return tx.Commit()
}

func toDomainPageVersion(row sqlcgen.PageVersion) domain.PageVersion {
	v := domain.PageVersion{
		PageID:       row.PageID.String(),
		Seq:          row.Seq,
		Doc:          string(row.Doc),
		AuthorUserID: uint64(row.AuthorUserID),
		CreatedAt:    row.CreatedAt,
	}
	if row.Note.Valid {
		note := row.Note.String
		v.Note = &note
	}
	return v
}

func (r *pageVersionRepository) CreateVersionIfDue(
	ctx context.Context,
	workspaceID, pageID string,
	doc string,
	authorUserID uint64,
	note *string,
	force bool,
) (bool, *domain.PageVersion, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return false, nil, repository.ErrPageNotFound
	}
	authorID, ok3 := toInt64ID(authorUserID)
	if !ok3 {
		return false, nil, outOfRangeIDError("author_user_id", authorUserID)
	}

	var created bool
	var version *domain.PageVersion
	err := r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		// 1. pages 行をロックする。同時実行の設計（このファイルのコメント参照）により、
		// 「本文保存経路」「版を残す」「復元」の 3 つの書き込み経路すべてがここを通ってから
		// 最大 seq を読むため、page_versions の PK (page_id, seq) 衝突は原理的に起きない。
		// 対象ページが無ければ 0 行（sql.ErrNoRows）で ErrPageNotFound に翻訳する。
		//
		// 既にこのトランザクションの中で pages 行をロック済み（例えば
		// ReplacePageBlocksUseCase の TouchPageLastEditedBy 経由）でも、PostgreSQL の行ロックは
		// 同一トランザクション内では再入可能なので、ここで自分自身のロック待ちにはならない。
		if _, err := qtx.LockPageForVersioning(ctx, sqlcgen.LockPageForVersioningParams{
			WorkspaceID: wsID, PageID: pgID,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return repository.ErrPageNotFound
			}
			return err
		}

		// この操作全体を代表する時刻を 1 回だけ決め、10 分規則の判定・挿入する行の
		// created_at・掃除の cutoff のすべてにこの同じ値を使う。PostgreSQL の now() は
		// トランザクション開始時刻で固定されるため、列の DEFAULT に任せると「Go 側の
		// time.Now()（10分規則の判定に使う）」と「DB 側の now()（created_at に入る値）」
		// という 2 つの時刻の出どころが混ざってしまう（CodeRabbit 指摘）。ロックを取った
		// 直後のこの瞬間を全工程の基準時刻にする。
		now := time.Now()

		// 2. 直近の版（seq 降順 1 件）。1 つも無ければ「間引く対象が無い」= 必ず切る。
		latest, err := qtx.GetLatestPageVersion(ctx, sqlcgen.GetLatestPageVersionParams{
			WorkspaceID: wsID, PageID: pgID,
		})
		hasLatest := true
		switch {
		case errors.Is(err, sql.ErrNoRows):
			hasLatest = false
		case err != nil:
			return err
		}

		// 3. 間引きの判定。force（「版を残す」・復元）、直近が無い、10 分規則のいずれかで切る。
		shouldCut := force || !hasLatest
		if !shouldCut && now.Sub(latest.CreatedAt) > pageVersionMinInterval {
			shouldCut = true
		}
		if !shouldCut {
			created, version = false, nil
			return nil
		}

		// 4. seq = 直近 + 1（無ければ 1）で 1 件挿入する。
		nextSeq := int64(1)
		if hasLatest {
			nextSeq = latest.Seq + 1
		}
		row, err := qtx.CreatePageVersion(ctx, sqlcgen.CreatePageVersionParams{
			WorkspaceID:  wsID,
			PageID:       pgID,
			Seq:          nextSeq,
			Doc:          json.RawMessage(doc),
			AuthorUserID: authorID,
			Note:         nullString(note),
			CreatedAt:    now,
		})
		if err != nil {
			return err
		}

		// 5. 挿入と同じトランザクションで 30 日より古い版を掃除する。今挿入した行の
		// created_at（= now）は cutoff より新しいため、この DELETE の対象にはならない。
		if err := qtx.DeleteOldPageVersions(ctx, sqlcgen.DeleteOldPageVersionsParams{
			WorkspaceID: wsID, PageID: pgID, Cutoff: now.Add(-pageVersionRetention),
		}); err != nil {
			return err
		}

		v := toDomainPageVersion(row)
		created, version = true, &v
		return nil
	})
	if err != nil {
		return false, nil, err
	}
	return created, version, nil
}

func (r *pageVersionRepository) LockPage(ctx context.Context, workspaceID, pageID string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	return r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		if _, err := qtx.LockPageForVersioning(ctx, sqlcgen.LockPageForVersioningParams{
			WorkspaceID: wsID, PageID: pgID,
		}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return repository.ErrPageNotFound
			}
			return err
		}
		return nil
	})
}

func (r *pageVersionRepository) ListVersions(ctx context.Context, workspaceID, pageID string) ([]domain.PageVersion, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, domain.ErrPageVersionNotFound
	}
	rows, err := r.queries(ctx).ListPageVersions(ctx, sqlcgen.ListPageVersionsParams{
		WorkspaceID: wsID, PageID: pgID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.PageVersion, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainPageVersion(row))
	}
	return out, nil
}

func (r *pageVersionRepository) GetVersion(ctx context.Context, workspaceID, pageID string, seq int64) (*domain.PageVersion, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, domain.ErrPageVersionNotFound
	}
	row, err := r.queries(ctx).GetPageVersion(ctx, sqlcgen.GetPageVersionParams{
		WorkspaceID: wsID, PageID: pgID, Seq: seq,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPageVersionNotFound
	}
	if err != nil {
		return nil, err
	}
	v := toDomainPageVersion(row)
	return &v, nil
}
