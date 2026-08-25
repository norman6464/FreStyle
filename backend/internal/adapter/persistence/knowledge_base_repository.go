package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// knowledgeBaseRepository は [repository.KnowledgeBaseRepository] の実装。
// ナレッジ基盤は GORM を通さない方針（スキーマの正本は infra/database/schema/knowledge_base.sql）
// のため、クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
// 複数テーブルにまたがる書き込み（ページ作成・移動・本文置き換え）は BeginTx で
// この層に閉じたトランザクションにする（usecase に *sql.Tx を漏らさない）。
type knowledgeBaseRepository struct {
	db *sql.DB
	q  *sqlcgen.Queries
}

// NewKnowledgeBaseRepository はナレッジ基盤の repository を組み立てる。
func NewKnowledgeBaseRepository(db *sql.DB) repository.KnowledgeBaseRepository {
	return &knowledgeBaseRepository{db: db, q: sqlcgen.New(db)}
}

// kbParseID は文字列 ID を uuid に変換する。不正な形式は「存在し得ない ID = not found」として
// 扱えるよう ok=false を返す（URL 由来の生文字列を DB エラーにしないため）。
func kbParseID(id string) (uuid.UUID, bool) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, false
	}
	return u, true
}

// kbNullID は NULL 可の親 ID（*string）を uuid.NullUUID へ変換する。
func kbNullID(id *string) (uuid.NullUUID, bool) {
	if id == nil {
		return uuid.NullUUID{}, true
	}
	u, ok := kbParseID(*id)
	if !ok {
		return uuid.NullUUID{}, false
	}
	return uuid.NullUUID{UUID: u, Valid: true}, true
}

// kbNewID は UUIDv7 を採番する。時系列で単調に増える（インデックス局所性が良い）うえ、
// ランダム部により URL は推測困難のまま。失敗は乱数源の故障なのでエラーで返す。
func kbNewID() (uuid.UUID, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("uuid v7 の採番に失敗: %w", err)
	}
	return id, nil
}

func toDomainWorkspace(row sqlcgen.Workspace) domain.Workspace {
	return domain.Workspace{
		ID:        row.ID.String(),
		Slug:      row.Slug,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func toDomainSpace(row sqlcgen.Space) domain.Space {
	return domain.Space{
		ID:          row.ID.String(),
		WorkspaceID: row.WorkspaceID.String(),
		Key:         row.Key,
		Name:        row.Name,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toDomainPage(row sqlcgen.Page) domain.Page {
	p := domain.Page{
		ID:              row.ID.String(),
		WorkspaceID:     row.WorkspaceID.String(),
		SpaceID:         row.SpaceID.String(),
		Position:        row.Position,
		Title:           row.Title,
		CreatedByUserID: uint64(row.CreatedByUserID),
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.ParentID.Valid {
		id := row.ParentID.UUID.String()
		p.ParentID = &id
	}
	if row.ArchivedAt.Valid {
		t := row.ArchivedAt.Time
		p.ArchivedAt = &t
	}
	return p
}

func toDomainBlock(row sqlcgen.Block) domain.Block {
	b := domain.Block{
		ID:          row.ID.String(),
		WorkspaceID: row.WorkspaceID.String(),
		PageID:      row.PageID.String(),
		Position:    row.Position,
		Type:        domain.BlockType(row.Type),
		Attrs:       string(row.Attrs),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.ParentID.Valid {
		id := row.ParentID.UUID.String()
		b.ParentID = &id
	}
	if row.Inline != nil {
		s := string(*row.Inline)
		b.Inline = &s
	}
	return b
}

func toDomainPageSnapshot(row sqlcgen.PageSnapshot) domain.PageSnapshot {
	return domain.PageSnapshot{
		PageID:  row.PageID.String(),
		Doc:     string(row.Doc),
		BuiltAt: row.BuiltAt,
	}
}

func (r *knowledgeBaseRepository) FindWorkspaceByID(ctx context.Context, workspaceID string) (*domain.Workspace, error) {
	id, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	row, err := r.q.GetWorkspaceByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrWorkspaceNotFound
	}
	if err != nil {
		return nil, err
	}
	ws := toDomainWorkspace(row)
	return &ws, nil
}

func (r *knowledgeBaseRepository) FindSpace(ctx context.Context, workspaceID, spaceID string) (*domain.Space, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return nil, repository.ErrSpaceNotFound
	}
	row, err := r.q.GetSpace(ctx, sqlcgen.GetSpaceParams{WorkspaceID: wsID, ID: spID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrSpaceNotFound
	}
	if err != nil {
		return nil, err
	}
	sp := toDomainSpace(row)
	return &sp, nil
}

func (r *knowledgeBaseRepository) FindPage(ctx context.Context, workspaceID, pageID string) (*domain.Page, error) {
	return findPageWith(ctx, r.q, workspaceID, pageID)
}

// findPageWith はトランザクション内外の両方から使うページ取得の実体。
func findPageWith(ctx context.Context, q *sqlcgen.Queries, workspaceID, pageID string) (*domain.Page, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	row, err := q.GetPage(ctx, sqlcgen.GetPageParams{WorkspaceID: wsID, ID: pgID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPageNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toDomainPage(row)
	return &p, nil
}

func (r *knowledgeBaseRepository) ListActivePagesBySpace(ctx context.Context, workspaceID, spaceID string) ([]domain.Page, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return []domain.Page{}, nil
	}
	rows, err := r.q.ListActivePagesBySpace(ctx, sqlcgen.ListActivePagesBySpaceParams{WorkspaceID: wsID, SpaceID: spID})
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		pages = append(pages, toDomainPage(row))
	}
	return pages, nil
}

func (r *knowledgeBaseRepository) LastActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string) (string, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	parent, ok3 := kbNullID(parentID)
	if !ok || !ok2 || !ok3 {
		return "", nil
	}
	pos, err := r.q.GetLastActiveSiblingPosition(ctx, sqlcgen.GetLastActiveSiblingPositionParams{
		WorkspaceID: wsID,
		SpaceID:     spID,
		ParentID:    parent,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil // 兄弟なし = fracindex.Between の「端」
	}
	if err != nil {
		return "", err
	}
	return pos, nil
}

func (r *knowledgeBaseRepository) HasActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string, position, excludePageID string) (bool, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	parent, ok3 := kbNullID(parentID)
	exID, ok4 := kbParseID(excludePageID)
	if !ok || !ok2 || !ok3 || !ok4 {
		return false, nil
	}
	return r.q.HasActiveSiblingPosition(ctx, sqlcgen.HasActiveSiblingPositionParams{
		WorkspaceID:    wsID,
		SpaceID:        spID,
		ParentID:       parent,
		Position:       position,
		ExcludedPageID: exID,
	})
}

func (r *knowledgeBaseRepository) HasDescendant(ctx context.Context, workspaceID, pageID, candidateID string) (bool, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	cdID, ok3 := kbParseID(candidateID)
	if !ok || !ok2 || !ok3 {
		return false, nil
	}
	return r.q.PageHasDescendant(ctx, sqlcgen.PageHasDescendantParams{
		WorkspaceID: wsID,
		AncestorID:  pgID,
		PageID:      cdID,
	})
}

func (r *knowledgeBaseRepository) CreatePage(ctx context.Context, page *domain.Page) error {
	wsID, ok := kbParseID(page.WorkspaceID)
	spID, ok2 := kbParseID(page.SpaceID)
	parent, ok3 := kbNullID(page.ParentID)
	if !ok || !ok2 || !ok3 {
		return repository.ErrPageNotFound
	}
	id, err := kbNewID()
	if err != nil {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op

	qtx := r.q.WithTx(tx)
	row, err := qtx.InsertPage(ctx, sqlcgen.InsertPageParams{
		ID:              id,
		WorkspaceID:     wsID,
		SpaceID:         spID,
		ParentID:        parent,
		Position:        page.Position,
		Title:           page.Title,
		CreatedByUserID: int64(page.CreatedByUserID),
	})
	if err != nil {
		return err
	}
	// closure: 自分自身（depth=0）と、親があれば親の祖先集合 +1。
	if err := qtx.InsertPagePathSelf(ctx, sqlcgen.InsertPagePathSelfParams{WorkspaceID: wsID, PageID: id}); err != nil {
		return err
	}
	if parent.Valid {
		if err := qtx.InsertPagePathAncestors(ctx, sqlcgen.InsertPagePathAncestorsParams{
			PageID:      id,
			WorkspaceID: wsID,
			ParentID:    parent.UUID,
		}); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	*page = toDomainPage(row)
	return nil
}

func (r *knowledgeBaseRepository) UpdatePageTitle(ctx context.Context, workspaceID, pageID, title string) (*domain.Page, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	row, err := r.q.UpdatePageTitle(ctx, sqlcgen.UpdatePageTitleParams{WorkspaceID: wsID, ID: pgID, Title: title})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPageNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toDomainPage(row)
	return &p, nil
}

func (r *knowledgeBaseRepository) MovePage(ctx context.Context, workspaceID, pageID string, newParentID *string, newSpaceID, newPosition string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	parent, ok3 := kbNullID(newParentID)
	spID, ok4 := kbParseID(newSpaceID)
	if !ok || !ok2 || !ok3 || !ok4 {
		return repository.ErrPageNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.q.WithTx(tx)
	current, err := findPageWith(ctx, qtx, workspaceID, pageID)
	if err != nil {
		return err
	}
	if current.SpaceID == newSpaceID {
		n, err := qtx.MovePageWithinSpace(ctx, sqlcgen.MovePageWithinSpaceParams{
			NewParentID: parent,
			NewPosition: newPosition,
			WorkspaceID: wsID,
			PageID:      pgID,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return repository.ErrPageNotFound
		}
	} else {
		// スペースをまたぐ移動は本人 + 子孫の space_id を 1 文で更新する（クエリ側コメント参照）。
		n, err := qtx.MovePageSubtreeToSpace(ctx, sqlcgen.MovePageSubtreeToSpaceParams{
			NewSpaceID:  spID,
			PageID:      pgID,
			NewParentID: parent,
			NewPosition: newPosition,
			WorkspaceID: wsID,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return repository.ErrPageNotFound
		}
	}
	// closure の付け替え: 旧祖先との組を消してから、新しい親の祖先集合との組を張る。
	// 順序は Detach → Attach 固定（逆にすると Attach で張った行を Detach が消してしまう）。
	if err := qtx.DetachPageSubtreePaths(ctx, sqlcgen.DetachPageSubtreePathsParams{
		WorkspaceID: wsID,
		PageID:      pgID,
	}); err != nil {
		return err
	}
	if parent.Valid {
		if err := qtx.AttachPageSubtreePaths(ctx, sqlcgen.AttachPageSubtreePathsParams{
			NewParentID: parent.UUID,
			WorkspaceID: wsID,
			PageID:      pgID,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *knowledgeBaseRepository) ArchivePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	// 1 文で完結する（サブツリー全行に同じ now() が入る）ためトランザクション不要。
	_, err := r.q.ArchivePageSubtree(ctx, sqlcgen.ArchivePageSubtreeParams{WorkspaceID: wsID, AncestorID: pgID})
	return err
}

func (r *knowledgeBaseRepository) UnarchivePageSubtree(ctx context.Context, workspaceID, pageID string, archivedSince time.Time, newRootPosition *string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.q.WithTx(tx)
	// 現役の兄弟と position が衝突する場合は、まだアーカイブ済み（部分 UNIQUE の対象外）の
	// うちに根の position を振り直してから現役へ戻す。
	if newRootPosition != nil {
		if _, err := qtx.SetPagePosition(ctx, sqlcgen.SetPagePositionParams{
			WorkspaceID: wsID,
			ID:          pgID,
			Position:    *newRootPosition,
		}); err != nil {
			return err
		}
	}
	if _, err := qtx.UnarchivePageSubtree(ctx, sqlcgen.UnarchivePageSubtreeParams{
		WorkspaceID:   wsID,
		ArchivedSince: sql.NullTime{Time: archivedSince, Valid: true},
		PageID:        pgID,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *knowledgeBaseRepository) ListBlocksByPage(ctx context.Context, workspaceID, pageID string) ([]domain.Block, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.Block{}, nil
	}
	rows, err := r.q.ListBlocksByPage(ctx, sqlcgen.ListBlocksByPageParams{WorkspaceID: wsID, PageID: pgID})
	if err != nil {
		return nil, err
	}
	blocks := make([]domain.Block, 0, len(rows))
	for _, row := range rows {
		blocks = append(blocks, toDomainBlock(row))
	}
	return blocks, nil
}

func (r *knowledgeBaseRepository) ReplacePageBlocks(ctx context.Context, workspaceID, pageID string, blocks []repository.BlockWrite, snapshotDoc string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	qtx := r.q.WithTx(tx)
	// page_snapshots は workspace_id を持たないため、同一トランザクション内で
	// ページの所属を必ず検証してから書く（テナント越えの snapshot 書き込みを塞ぐ）。
	if _, err := findPageWith(ctx, qtx, workspaceID, pageID); err != nil {
		return err
	}
	if err := qtx.DeletePageBlocks(ctx, sqlcgen.DeletePageBlocksParams{WorkspaceID: wsID, PageID: pgID}); err != nil {
		return err
	}
	ids := make([]uuid.UUID, len(blocks))
	for i, b := range blocks {
		id, err := kbNewID()
		if err != nil {
			return err
		}
		ids[i] = id
		var parent uuid.NullUUID
		if b.ParentIndex >= 0 {
			// 文書順（親が先）が前提。壊れた入力で別ページの行を親にしないよう添字を検証する。
			if b.ParentIndex >= i {
				return fmt.Errorf("blocks[%d] の ParentIndex %d が自分より後を指しています", i, b.ParentIndex)
			}
			parent = uuid.NullUUID{UUID: ids[b.ParentIndex], Valid: true}
		}
		var inline *json.RawMessage
		if b.Inline != nil {
			raw := json.RawMessage(*b.Inline)
			inline = &raw
		}
		if err := qtx.InsertBlock(ctx, sqlcgen.InsertBlockParams{
			ID:          id,
			WorkspaceID: wsID,
			PageID:      pgID,
			ParentID:    parent,
			Position:    b.Position,
			Type:        string(b.Type),
			Attrs:       json.RawMessage(b.Attrs),
			Inline:      inline,
		}); err != nil {
			return err
		}
	}
	if err := qtx.UpsertPageSnapshot(ctx, sqlcgen.UpsertPageSnapshotParams{
		PageID: pgID,
		Doc:    json.RawMessage(snapshotDoc),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *knowledgeBaseRepository) GetPageSnapshot(ctx context.Context, workspaceID, pageID string) (*domain.PageSnapshot, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageSnapshotNotFound
	}
	row, err := r.q.GetPageSnapshot(ctx, sqlcgen.GetPageSnapshotParams{WorkspaceID: wsID, PageID: pgID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPageSnapshotNotFound
	}
	if err != nil {
		return nil, err
	}
	s := toDomainPageSnapshot(row)
	return &s, nil
}
