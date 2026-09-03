package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// courseRepository は [repository.CourseRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type courseRepository struct {
	db *sql.DB
}

func NewCourseRepository(db *sql.DB) repository.CourseRepository {
	return &courseRepository{db: db}
}

func toDomainCourse(row sqlcgen.Course) domain.Course {
	c := domain.Course{
		ID:              uint64(row.ID),
		CreatedByUserID: uint64(row.CreatedByUserID),
		Title:           row.Title,
		Description:     row.Description,
		Category:        row.Category,
		Language:        row.Language,
		SortOrder:       int(row.SortOrder),
		IsPublished:     row.IsPublished,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		c.WorkspaceID = &wid
	}
	return c
}

// ListByCompany は自社のコースを sort_order 昇順で返す。includeUnpublished=false なら公開のみ。
func (r *courseRepository) ListByWorkspaceID(ctx context.Context, workspaceID string, includeUnpublished bool) ([]domain.Course, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return []domain.Course{}, nil // 不正 / 空の ID は該当なしと同じ扱い
	}
	rows, err := sqlcgen.New(r.db).ListCoursesByWorkspace(ctx, sqlcgen.ListCoursesByWorkspaceParams{
		WorkspaceID:        wid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	courses := make([]domain.Course, 0, len(rows))
	for _, row := range rows {
		courses = append(courses, toDomainCourse(row))
	}
	return courses, nil
}

// GetByID は単一コースを返す。未存在は domain.ErrNotFound（handler が 404 に分岐）。
func (r *courseRepository) GetByID(ctx context.Context, id uint64) (*domain.Course, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	row, err := sqlcgen.New(r.db).GetCourseByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	c := toDomainCourse(row)
	return &c, nil
}

func (r *courseRepository) Create(ctx context.Context, c *domain.Course) error {
	return insertCourseWith(ctx, sqlcgen.New(r.db), c)
}

// insertCourseWith はコースを 1 行入れて、採番された値を c へ書き戻す。
// 単独の Create とトランザクション版の両方から通す（組み立てを 2 か所に書かない）。
func insertCourseWith(ctx context.Context, q *sqlcgen.Queries, c *domain.Course) error {
	wid, ok := nullWorkspaceID(c.WorkspaceID)
	if !ok {
		return fmt.Errorf("workspace_id が不正な形式です: %q", *c.WorkspaceID)
	}
	createdBy, ok := toInt64ID(c.CreatedByUserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("created_by", c.CreatedByUserID)
	}
	now := time.Now()
	createdAt := c.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := c.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}
	row, err := q.InsertCourse(ctx, sqlcgen.InsertCourseParams{
		WorkspaceID:     wid,
		CreatedByUserID: createdBy,
		Title:           c.Title,
		Description:     c.Description,
		Category:        c.Category,
		Language:        c.Language,
		SortOrder:       int64(c.SortOrder), // 0 は SQL 側の COALESCE で既定 100 に倒す
		IsPublished:     c.IsPublished,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	})
	if err != nil {
		return err
	}
	c.ID = uint64(row.ID)
	c.SortOrder = int(row.SortOrder) // 既定 100 が当たった場合を書き戻す（GORM の default タグ相当）
	c.CreatedAt = row.CreatedAt
	c.UpdatedAt = row.UpdatedAt
	return nil
}

// CreateWithOwnerGrant はコースと、作成者への admin の付与を 1 つのトランザクションで書く。
//
// 分けて書くと、間に落ちたときに「誰も編集できないコース」が残る。いまのワークスペースには
// admin が居ないので、そうなると作り直す以外に直す手が無い。
func (r *courseRepository) CreateWithOwnerGrant(ctx context.Context, c *domain.Course, ownerPrincipalID string) error {
	prID, ok := kbParseID(ownerPrincipalID)
	if !ok {
		return fmt.Errorf("principal_id が不正な形式です: %q", ownerPrincipalID)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	q := sqlcgen.New(tx)
	if err := insertCourseWith(ctx, q, c); err != nil {
		return err
	}
	courseID, ok := toInt64ID(c.ID)
	if !ok {
		return outOfRangeIDError("course", c.ID)
	}
	wid, _ := nullWorkspaceID(c.WorkspaceID)
	if _, err := q.UpsertCourseGrant(ctx, sqlcgen.UpsertCourseGrantParams{
		WorkspaceID: wid.UUID,
		CourseID:    courseID,
		PrincipalID: prID,
		Role:        string(domain.GrantRoleAdmin),
	}); err != nil {
		return err
	}
	return tx.Commit()
}

// Update はコースの 4 列を書き換える。対象行が無ければ domain.ErrNotFound を返す。
//

func (r *courseRepository) Update(ctx context.Context, c *domain.Course) error {
	id64, ok := toInt64ID(c.ID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// CreatedBy / Category / Language は更新対象外（GORM の Updates(map) と同じ 4 列のみ）。
	updatedAt, err := sqlcgen.New(r.db).UpdateCourse(ctx, sqlcgen.UpdateCourseParams{
		ID:          id64,
		Title:       c.Title,
		Description: c.Description,
		SortOrder:   int64(c.SortOrder),
		IsPublished: c.IsPublished,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound // 0 行更新 = 取得と更新のあいだにコースが消えた
	}
	if err != nil {
		return err
	}
	c.UpdatedAt = updatedAt
	return nil
}

// Delete はコースを物理削除する。対象行が無ければ domain.ErrNotFound を返す。
//

func (r *courseRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// :execrows なので実際に消えた行数が返る（:exec だと 0 行でも成功と区別が付かない）。
	affected, err := sqlcgen.New(r.db).DeleteCourse(ctx, id64)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
