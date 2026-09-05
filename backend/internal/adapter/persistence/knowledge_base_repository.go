package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// knowledgeBaseRepository は [repository.KnowledgeBaseRepository] の実装。
// ノートは GORM を通さない方針（スキーマの正本は infra/database/schema/knowledge_base.sql）
// のため、クエリはすべて sqlc 生成コード + 素の *sql.DB で書く。
// 複数テーブルにまたがる書き込み（ページ作成・移動・本文置き換え）は BeginTx で
// この層に閉じたトランザクションにする（usecase に *sql.Tx を漏らさない）。
type knowledgeBaseRepository struct {
	baseRepository
}

// NewKnowledgeBaseRepository はノートの repository を組み立てる。
func NewKnowledgeBaseRepository(db *sql.DB) repository.KnowledgeBaseRepository {
	return &knowledgeBaseRepository{baseRepository{db: db}}
}

// queries は ctx に乗っているトランザクション（あれば）に束縛した sqlc の Queries を作る。
func (r *knowledgeBaseRepository) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(r.dbtx(ctx))
}

// runInTx は 1 つのトランザクションを開き、その中でだけ有効な Queries を fn に渡す。
// ctx に既に外側の DoInTx が開いたトランザクションがあれば、新規に開始せずそれへ相乗りする
// （二重に BeginTx するとデッドロックの原因になる。commit/rollback は外側だけが持つ）。
func (r *knowledgeBaseRepository) runInTx(ctx context.Context, fn func(qtx *sqlcgen.Queries) error) error {
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

// PostgreSQL の SQLSTATE。制約違反を「業務上の衝突」へ翻訳するのに使う。
const (
	sqlStateUniqueViolation     = "23505"
	sqlStateForeignKeyViolation = "23503"
)

// isUniqueViolation は一意制約違反（重複）かを返す。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == sqlStateUniqueViolation
}

// uniqueViolationConstraint は一意制約違反のとき、違反した制約名を返す。
// 1 つの INSERT が複数の一意制約を持ちうる場合、名前を見ないとどの制約が競合したか
// 区別できず、意味の違うエラー（本当に重複 / 別の要求が既に作っていた等）を
// 取り違えて返してしまう。
func uniqueViolationConstraint(err error) (string, bool) {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != sqlStateUniqueViolation {
		return "", false
	}
	return pgErr.ConstraintName, true
}

// isForeignKeyViolation は外部キー違反（参照先が無い）かを返す。
func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == sqlStateForeignKeyViolation
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
		IsActive:  row.IsActive,
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
		Visibility:  domain.SpaceVisibility(row.Visibility),
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

// DeleteWorkspace はワークスペースを配下ごと消す。
//
// 0 行だったときに「無かった」と「人が居るから消さなかった」を撃ち分ける必要がある。
// SQL 側は人が居るものを WHERE で弾くだけなのでどちらも 0 行になる。ここで実在を引き直し、
// 在るのに消えなかった＝人が居ると判定する（守りは SQL 側にあり、ここは理由付けだけ）。
func (r *knowledgeBaseRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	id, ok := kbParseID(workspaceID)
	if !ok {
		return repository.ErrWorkspaceNotFound
	}
	affected, err := r.queries(ctx).DeleteWorkspace(ctx, id)
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}
	if _, err := r.queries(ctx).GetWorkspaceByID(ctx, id); errors.Is(err, sql.ErrNoRows) {
		return repository.ErrWorkspaceNotFound
	} else if err != nil {
		return err
	}
	return repository.ErrWorkspaceHasMembers
}

func (r *knowledgeBaseRepository) FindWorkspaceByID(ctx context.Context, workspaceID string) (*domain.Workspace, error) {
	id, ok := kbParseID(workspaceID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	row, err := r.queries(ctx).GetWorkspaceByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrWorkspaceNotFound
	}
	if err != nil {
		return nil, err
	}
	ws := toDomainWorkspace(row)
	return &ws, nil
}

func (r *knowledgeBaseRepository) FindPersonalWorkspaceByOwner(ctx context.Context, userID uint64) (*domain.Workspace, error) {
	id, ok := toInt64ID(userID)
	if !ok {
		return nil, repository.ErrWorkspaceNotFound
	}
	row, err := r.queries(ctx).GetPersonalWorkspaceByOwner(ctx, sql.NullInt64{Int64: id, Valid: true})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrWorkspaceNotFound
	}
	if err != nil {
		return nil, err
	}
	ws := toDomainWorkspace(row)
	return &ws, nil
}

func (r *knowledgeBaseRepository) FindWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	if slug == "" {
		return nil, repository.ErrWorkspaceNotFound
	}
	row, err := r.queries(ctx).GetWorkspaceBySlug(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrWorkspaceNotFound
	}
	if err != nil {
		return nil, err
	}
	ws := toDomainWorkspace(row)
	return &ws, nil
}

// FindPageByIDAcrossWorkspaces はページを ID だけで引く（詳細は port のコメント）。
func (r *knowledgeBaseRepository) FindPageByIDAcrossWorkspaces(ctx context.Context, pageID string) (*domain.Page, error) {
	pgID, ok := kbParseID(pageID)
	if !ok {
		return nil, repository.ErrPageNotFound
	}
	row, err := r.queries(ctx).GetPageAcrossWorkspaces(ctx, pgID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPageNotFound
	}
	if err != nil {
		return nil, err
	}
	p := toDomainPage(row)
	return &p, nil
}

func (r *knowledgeBaseRepository) ListAncestorPageIDs(ctx context.Context, workspaceID, pageID string) ([]string, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []string{}, nil
	}
	rows, err := r.queries(ctx).ListPageAncestorIDs(ctx, sqlcgen.ListPageAncestorIDsParams{
		WorkspaceID: wsID,
		PageID:      pgID,
	})
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, id := range rows {
		out = append(out, id.String())
	}
	return out, nil
}

func (r *knowledgeBaseRepository) DeletePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	rows, err := r.queries(ctx).DeletePage(ctx, sqlcgen.DeletePageParams{WorkspaceID: wsID, ID: pgID})
	if err != nil {
		return err
	}
	if rows == 0 {
		return repository.ErrPageNotFound
	}
	return nil
}

func (r *knowledgeBaseRepository) FindSpace(ctx context.Context, workspaceID, spaceID string) (*domain.Space, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return nil, repository.ErrSpaceNotFound
	}
	row, err := r.queries(ctx).GetSpace(ctx, sqlcgen.GetSpaceParams{WorkspaceID: wsID, ID: spID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrSpaceNotFound
	}
	if err != nil {
		return nil, err
	}
	sp := toDomainSpace(row)
	return &sp, nil
}

// UpdateSpaceName はスペースの表示名を変える。0 件更新は「無い」と同じ扱いで
// ErrSpaceNotFound（別ワークスペースの ID も WHERE の workspace_id でここに落ちる）。
func (r *knowledgeBaseRepository) UpdateSpaceName(ctx context.Context, workspaceID, spaceID, name string) error {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	if !ok || !ok2 {
		return repository.ErrSpaceNotFound
	}
	affected, err := r.queries(ctx).UpdateSpaceName(ctx, sqlcgen.UpdateSpaceNameParams{
		WorkspaceID: wsID,
		ID:          spID,
		Name:        name,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return repository.ErrSpaceNotFound
	}
	return nil
}

func (r *knowledgeBaseRepository) CreateSpace(ctx context.Context, space *domain.Space) error {
	wsID, ok := kbParseID(space.WorkspaceID)
	if !ok {
		return repository.ErrWorkspaceNotFound
	}
	id, err := kbNewID()
	if err != nil {
		return err
	}
	// ゼロ値（visibility 未指定の呼び出し）は既定の 'workspace' に倒す。
	// 空文字のまま送ると CHECK 制約（ck_spaces_visibility）で落ちる。
	visibility := space.Visibility
	if visibility == "" {
		visibility = domain.SpaceVisibilityWorkspace
	}
	row, err := r.queries(ctx).InsertSpace(ctx, sqlcgen.InsertSpaceParams{
		ID:          id,
		WorkspaceID: wsID,
		Key:         space.Key,
		Name:        space.Name,
		Visibility:  string(visibility),
	})
	if err != nil {
		// key の重複（uq_spaces_workspace_key）は入口の検証では防げない
		// （検査してから INSERT するまでの間に別の要求が同じ key を取り得る）。
		// 一意制約を唯一の判定にして、業務上の衝突として返す。
		if isUniqueViolation(err) {
			return repository.ErrSpaceKeyTaken
		}
		// ワークスペースが実在しなければ FK 違反になる。存在しないテナントへの
		// 作成要求なので「無い」に翻訳する（FK 違反を 500 で返すと、
		// クライアントは再試行してよいものと誤解する）。
		if isForeignKeyViolation(err) {
			return repository.ErrWorkspaceNotFound
		}
		return err
	}
	*space = toDomainSpace(row)
	return nil
}

func (r *knowledgeBaseRepository) FindPage(ctx context.Context, workspaceID, pageID string) (*domain.Page, error) {
	return findPageWith(ctx, r.queries(ctx), workspaceID, pageID)
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
	rows, err := r.queries(ctx).ListActivePagesBySpace(ctx, sqlcgen.ListActivePagesBySpaceParams{WorkspaceID: wsID, SpaceID: spID})
	if err != nil {
		return nil, err
	}
	pages := make([]domain.Page, 0, len(rows))
	for _, row := range rows {
		pages = append(pages, toDomainPage(row))
	}
	return pages, nil
}

func (r *knowledgeBaseRepository) SiblingPositionsAround(
	ctx context.Context, workspaceID, spaceID string, parentID *string, anchorPageID, movingPageID string,
) (bool, string, string, string, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	parent, ok3 := kbNullID(parentID)
	anchorID, ok4 := kbParseID(anchorPageID)
	// movingPageID が空なら「除くものが無い」。どのページの ID とも一致しない値を渡して、
	// 除外の条件をそのまま無効化する（条件を組み替えて分岐を増やさない）。
	movingID := uuid.Nil
	ok5 := true
	if movingPageID != "" {
		movingID, ok5 = kbParseID(movingPageID)
	}
	if !ok || !ok2 || !ok3 || !ok4 || !ok5 {
		// UUID ですらない値はどの兄弟にも一致しない。**エラーにはしない**
		// （見つからなかったときと同じ扱い。撃ち分けると応答の作られ方が変わる）。
		return false, "", "", "", nil
	}
	row, err := r.queries(ctx).SiblingPositionsAround(ctx, sqlcgen.SiblingPositionsAroundParams{
		WorkspaceID:  wsID,
		SpaceID:      spID,
		ParentID:     parent,
		AnchorPageID: anchorID,
		MovingPageID: movingID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return false, "", "", "", nil
	}
	if err != nil {
		return false, "", "", "", err
	}
	return row.Found, row.PrevPosition, row.AnchorPosition, row.NextPosition, nil
}

func (r *knowledgeBaseRepository) LastActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string) (string, error) {
	wsID, ok := kbParseID(workspaceID)
	spID, ok2 := kbParseID(spaceID)
	parent, ok3 := kbNullID(parentID)
	if !ok || !ok2 || !ok3 {
		return "", nil
	}
	pos, err := r.queries(ctx).GetLastActiveSiblingPosition(ctx, sqlcgen.GetLastActiveSiblingPositionParams{
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
	return r.queries(ctx).HasActiveSiblingPosition(ctx, sqlcgen.HasActiveSiblingPositionParams{
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
	return r.queries(ctx).PageHasDescendant(ctx, sqlcgen.PageHasDescendantParams{
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
	// pages.created_by_user_id は bigint（int64）で、domain のユーザー ID は uint64。
	// int64(page.CreatedByUserID) と素で書くと math.MaxInt64 を超える値が負数へ巻き戻り、
	// 作成者とは無関係な id が作成者として記録される。範囲外の id を持つユーザーは
	// 存在し得ないので、書き込みに入る前にエラーで止める
	// （nil を返すとページを作れたと誤認され、呼び出し側が page.ID を読みに行く）。
	createdBy, ok4 := toInt64ID(page.CreatedByUserID)
	if !ok4 {
		return outOfRangeIDError("created_by_user_id", page.CreatedByUserID)
	}
	id, err := kbNewID()
	if err != nil {
		return err
	}

	var created sqlcgen.Page
	err = r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		row, err := qtx.InsertPage(ctx, sqlcgen.InsertPageParams{
			ID:              id,
			WorkspaceID:     wsID,
			SpaceID:         spID,
			ParentID:        parent,
			Position:        page.Position,
			Title:           page.Title,
			CreatedByUserID: createdBy,
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
		created = row
		return nil
	})
	if err != nil {
		return err
	}
	*page = toDomainPage(created)
	return nil
}

func (r *knowledgeBaseRepository) UpdatePageTitle(ctx context.Context, workspaceID, pageID, title string) (*domain.Page, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageNotFound
	}
	row, err := r.queries(ctx).UpdatePageTitle(ctx, sqlcgen.UpdatePageTitleParams{WorkspaceID: wsID, ID: pgID, Title: title})
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

	return r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
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
			// 「そのスペースの全員」宛ての付与はスペースをまたぐと評価されなくなり、行は
			// 権限設定画面に見えているのに効かない状態になる。同じトランザクションで
			// 調べて拒否する（先に調べても、移動までのあいだに張られた行を取りこぼす）。
			voids, err := qtx.SubtreeHasForeignSpaceAllGrant(ctx, sqlcgen.SubtreeHasForeignSpaceAllGrantParams{
				WorkspaceID: wsID,
				PageID:      pgID,
				NewSpaceID:  spID,
			})
			if err != nil {
				return err
			}
			if voids {
				return repository.ErrPageMoveVoidsSpaceGrant
			}
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
		return nil
	})
}

// ArchivePageSubtree は根とその子孫をまとめてアーカイブする。
// 1 行も畳めなかった場合は repository.ErrPageNotFound を返す（handler が 404 にマップ）。
//
// 0 行更新を成功にしてはいけない理由:
//
//	UPDATE は 1 行も一致しなくても SQL としては成功する。ここで件数を捨てて nil を返すと
//	handler は 204 を返し、ツリーから消えたはずのページがそのまま残る。
//	この文は根も含めて畳むので、成功したなら必ず 1 行以上（= 根の分）に当たる。
//	0 行は「そのページがワークスペースに無い」ことしか意味しない。
//	呼び出し側（ArchivePageUseCase）は FindPage で存在を先に確かめているので、
//	ここに落ちるのは「確認とアーカイブのあいだにページが消えた」競合のときだけ。
func (r *knowledgeBaseRepository) ArchivePageSubtree(ctx context.Context, workspaceID, pageID string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}
	// 1 文で完結する（サブツリー全行に同じ now() が入る）ためトランザクション不要。
	// :execrows なので実際に畳んだ行数が返る（捨てると 0 行でも成功と区別が付かない）。
	n, err := r.queries(ctx).ArchivePageSubtree(ctx, sqlcgen.ArchivePageSubtreeParams{WorkspaceID: wsID, AncestorID: pgID})
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrPageNotFound
	}
	return nil
}

func (r *knowledgeBaseRepository) UnarchivePageSubtree(ctx context.Context, workspaceID, pageID string, archivedSince time.Time, newRootPosition *string) error {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return repository.ErrPageNotFound
	}

	return r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
		// 現役の兄弟と position が衝突する場合は、まだアーカイブ済み（部分 UNIQUE の対象外）の
		// うちに根の position を振り直してから現役へ戻す。
		//
		// ここも件数を見る。振り直しが 0 行なら根のページが無いということで、そのまま
		// UnarchivePageSubtree へ進むと「元の position のまま復帰させようとして UNIQUE で落ちる」か
		// 「何も起きないまま成功を返す」かのどちらかになり、どちらも原因が分からない形で表に出る。
		if newRootPosition != nil {
			n, err := qtx.SetPagePosition(ctx, sqlcgen.SetPagePositionParams{
				WorkspaceID: wsID,
				ID:          pgID,
				Position:    *newRootPosition,
			})
			if err != nil {
				return err
			}
			if n == 0 {
				return repository.ErrPageNotFound
			}
		}
		// 根も含めて戻す文なので、成功したなら必ず 1 行以上に当たる。0 行のまま成功を返すと
		// handler が 200 を返し、アーカイブされたままのページを復帰済みとして描画してしまう。
		n, err := qtx.UnarchivePageSubtree(ctx, sqlcgen.UnarchivePageSubtreeParams{
			WorkspaceID:   wsID,
			ArchivedSince: sql.NullTime{Time: archivedSince, Valid: true},
			PageID:        pgID,
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return repository.ErrPageNotFound
		}
		return nil
	})
}

func (r *knowledgeBaseRepository) ListBlocksByPage(ctx context.Context, workspaceID, pageID string) ([]domain.Block, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return []domain.Block{}, nil
	}
	rows, err := r.queries(ctx).ListBlocksByPage(ctx, sqlcgen.ListBlocksByPageParams{WorkspaceID: wsID, PageID: pgID})
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

	return r.runInTx(ctx, func(qtx *sqlcgen.Queries) error {
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
		return qtx.UpsertPageSnapshot(ctx, sqlcgen.UpsertPageSnapshotParams{
			PageID: pgID,
			Doc:    json.RawMessage(snapshotDoc),
		})
	})
}

func (r *knowledgeBaseRepository) GetPageSnapshot(ctx context.Context, workspaceID, pageID string) (*domain.PageSnapshot, error) {
	wsID, ok := kbParseID(workspaceID)
	pgID, ok2 := kbParseID(pageID)
	if !ok || !ok2 {
		return nil, repository.ErrPageSnapshotNotFound
	}
	row, err := r.queries(ctx).GetPageSnapshot(ctx, sqlcgen.GetPageSnapshotParams{WorkspaceID: wsID, PageID: pgID})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrPageSnapshotNotFound
	}
	if err != nil {
		return nil, err
	}
	s := toDomainPageSnapshot(row)
	return &s, nil
}
