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
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := c.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertCourse(ctx, sqlcgen.InsertCourseParams{
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

// Update はコースの 4 列を書き換える。対象行が無ければ domain.ErrNotFound を返す。
//
// 0 行更新を成功にしてはいけない理由:
//
//	UPDATE は 1 行も一致しなくても SQL としては成功する。ここで nil を返すと usecase は
//	「更新できた」と判断し、handler は 200 + 更新後のつもりの JSON を返す。利用者の画面には
//	保存済みと表示されるのに DB には何も書かれていない、という取り違えが起きる。
//	行が無いことは「保存できた」ではなく「対象が無い」なので 404 として伝える。
//
// UpdateCourse は :one（RETURNING updated_at）なので、0 行更新は sql.ErrNoRows として返る。
// 呼び出し側（CourseUseCase.Update）は GetByID で存在と権限を先に確かめているため、
// 実際にここへ落ちるのは「確認と更新のあいだにコースが消えた」競合のときだけ。
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
	c.UpdatedAt = updatedAt // GORM Save 相当の書き戻し
	return nil
}

// Delete はコースを物理削除する。対象行が無ければ domain.ErrNotFound を返す。
//
// DELETE でも 0 行を成功にしない理由:
//
//	「消えている」という事後条件だけを見れば 0 行削除も満たしている。それでもここで
//	not-found を返すのは、コース削除が CourseUseCase.Delete の中で
//	「配下教材の削除 → コース本体の削除」という 2 手に分かれているため。
//	本体が 1 行も消えなかったのに成功を返すと、教材だけ消してコースが残る状態
//	（= 空のコースが一覧に残る）を、呼び出し側が検知できないまま 204 で返してしまう。
//	usecase は GetByID で存在を先に確かめているので、ここに落ちるのは競合のときだけ。
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
