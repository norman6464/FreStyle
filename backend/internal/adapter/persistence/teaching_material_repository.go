package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// teachingMaterialRepository は [repository.TeachingMaterialRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type teachingMaterialRepository struct {
	db *gorm.DB
}

func NewTeachingMaterialRepository(db *gorm.DB) repository.TeachingMaterialRepository {
	return &teachingMaterialRepository{db: db}
}

// chapterDocPtr は NULL 可の doc(jsonb) を domain の *string へ写す（未移行の章は nil）。
func chapterDocPtr(raw *json.RawMessage) *string {
	if raw == nil {
		return nil
	}
	s := string(*raw)
	return &s
}

// toDomainChapter は行全体（本文 doc 含む）を domain へ写す。
func toDomainChapter(row sqlcgen.CourseChapter) domain.TeachingMaterial {
	return domain.TeachingMaterial{
		ID:              uint64(row.ID),
		CompanyID:       uint64(row.CompanyID),
		CourseID:        uint64(row.CourseID),
		CreatedByUserID: uint64(row.CreatedByUserID),
		Title:           row.Title,
		Doc:             chapterDocPtr(row.Doc),
		Revision:        int(row.Revision),
		SchemaVersion:   int(row.SchemaVersion),
		OrderInCourse:   int(row.SortOrder),
		IsPublished:     row.IsPublished,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// toDomainChapterSummary は一覧用の軽量行（本文 doc を含まない）を domain へ写す。Doc は nil のまま。
func toDomainChapterSummary(row sqlcgen.ListChaptersByCourseRow) domain.TeachingMaterial {
	return domain.TeachingMaterial{
		ID:              uint64(row.ID),
		CompanyID:       uint64(row.CompanyID),
		CourseID:        uint64(row.CourseID),
		CreatedByUserID: uint64(row.CreatedByUserID),
		Title:           row.Title,
		OrderInCourse:   int(row.SortOrder),
		IsPublished:     row.IsPublished,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

// ListByCompany は backward-compat 用（コース対応完了後に削除予定）。
func (r *teachingMaterialRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		return []domain.TeachingMaterial{}, nil // 存在し得ない company_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListChaptersByCompany(ctx, sqlcgen.ListChaptersByCompanyParams{
		CompanyID:          cid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	materials := make([]domain.TeachingMaterial, 0, len(rows))
	for _, row := range rows {
		materials = append(materials, toDomainChapter(row))
	}
	return materials, nil
}

// ListByCourse はコース内の章を sort_order 昇順で返す。
func (r *teachingMaterialRepository) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	cid, ok := toInt64ID(courseID)
	if !ok {
		return []domain.TeachingMaterial{}, nil // 存在し得ない course_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	// 一覧は本文（doc・jsonb）を返さない（章ごとに重く、全章を先読みすると非効率）。
	// 本文は選択時に GetByID で都度取得する。Doc は nil のままになる。
	rows, err := sqlcgen.New(sqlDB).ListChaptersByCourse(ctx, sqlcgen.ListChaptersByCourseParams{
		CourseID:           cid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	materials := make([]domain.TeachingMaterial, 0, len(rows))
	for _, row := range rows {
		materials = append(materials, toDomainChapterSummary(row))
	}
	return materials, nil
}

// GetByID は単一教材を返す。未存在は gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetChapterByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gorm.ErrRecordNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	m := toDomainChapter(row)
	return &m, nil
}

// CountByCourseForCompany は course_id ごとの教材件数を 1 クエリで集計する。
// trainee 向け(includeUnpublished=false)は published のみ数え、コース詳細の進捗分母と一致させる。
func (r *teachingMaterialRepository) CountByCourseForCompany(ctx context.Context, companyID uint64, includeUnpublished bool) (map[uint64]int, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		return map[uint64]int{}, nil // 存在し得ない company_id = 空
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).CountChaptersByCourseForCompany(ctx, sqlcgen.CountChaptersByCourseForCompanyParams{
		CompanyID:          cid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	counts := make(map[uint64]int, len(rows))
	for _, row := range rows {
		counts[uint64(row.CourseID)] = int(row.Cnt)
	}
	return counts, nil
}

func (r *teachingMaterialRepository) Create(ctx context.Context, m *domain.TeachingMaterial) error {
	companyID, ok := toInt64ID(m.CompanyID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("company_id", m.CompanyID)
	}
	courseID, ok := toInt64ID(m.CourseID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("course_id", m.CourseID)
	}
	createdBy, ok := toInt64ID(m.CreatedByUserID)
	if !ok {
		// 1 行も書けていないので nil を返さない（呼び出し側が作成できたと誤認する）。
		return outOfRangeIDError("created_by", m.CreatedByUserID)
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	now := time.Now()
	createdAt := m.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := m.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(sqlDB).InsertChapter(ctx, sqlcgen.InsertChapterParams{
		CompanyID:       companyID,
		CourseID:        courseID,
		CreatedByUserID: createdBy,
		Title:           m.Title,
		Revision:        int32(m.Revision),      // 0 は SQL 側の COALESCE で既定 1 に倒す
		SchemaVersion:   int32(m.SchemaVersion), // 0 は SQL 側の COALESCE で既定 1 に倒す
		SortOrder:       int32(m.OrderInCourse), // 0 は SQL 側の COALESCE で既定 100 に倒す
		IsPublished:     m.IsPublished,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	})
	if err != nil {
		return err
	}
	m.ID = uint64(row.ID)
	m.Revision = int(row.Revision)           // 既定 1 が当たった場合を書き戻す
	m.SchemaVersion = int(row.SchemaVersion) // 既定 1 が当たった場合を書き戻す
	m.OrderInCourse = int(row.SortOrder)     // 既定 100 が当たった場合を書き戻す
	m.CreatedAt = row.CreatedAt
	m.UpdatedAt = row.UpdatedAt
	return nil
}

// Update は title / sort_order / is_published を書き換える。対象行が無ければ
// gorm.ErrRecordNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) Update(ctx context.Context, m *domain.TeachingMaterial) error {
	id64, ok := toInt64ID(m.ID)
	if !ok {
		return gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	// CreatedBy / CompanyID / CourseID / Doc / Revision は不変（GORM の Updates(map) と同じ 3 列のみ）。
	// updated_at は now() へ進めて RETURNING で書き戻す（autoUpdateTime 相当）。
	updatedAt, err := sqlcgen.New(sqlDB).UpdateChapter(ctx, sqlcgen.UpdateChapterParams{
		ID:          id64,
		Title:       m.Title,
		SortOrder:   int64(m.OrderInCourse),
		IsPublished: m.IsPublished,
	})
	if err != nil {
		// 0 行 = 取得と更新の間に章が消えた。黙って nil を返すと失われた編集を保存済みに見せるので 404 を返す。
		if errors.Is(err, sql.ErrNoRows) {
			return gorm.ErrRecordNotFound
		}
		return err
	}
	m.UpdatedAt = updatedAt // GORM Save 相当の書き戻し
	return nil
}

// UpdateDocWithRevision はリッチ本文（tiptap JSON）を revision 一致条件の楽観ロックで更新する。
// rich_documents の UpdateWithRevision と同じパターン（0 行更新は存在確認で 404/409 を切り分け）。
func (r *teachingMaterialRepository) UpdateDocWithRevision(ctx context.Context, id uint64, doc string, expectedRevision int) (*domain.TeachingMaterial, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, gorm.ErrRecordNotFound // 存在し得ない id = not found
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(doc)
	row, err := sqlcgen.New(sqlDB).UpdateChapterDocWithRevision(ctx, sqlcgen.UpdateChapterDocWithRevisionParams{
		ID:               id64,
		Doc:              &raw,
		ExpectedRevision: int64(expectedRevision),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 0 行 = 「存在しない」か「版不一致」。存在確認で切り分ける。
			if _, gerr := r.GetByID(ctx, id); gerr != nil {
				return nil, gerr // gorm.ErrRecordNotFound
			}
			return nil, repository.ErrChapterDocConflict
		}
		return nil, mapChapterDocError(err)
	}
	m := toDomainChapter(row)
	return &m, nil
}

// mapChapterDocError は PostgreSQL の data exception（SQLSTATE class 22。jsonb に格納できない
// U+0000 等）を repository.ErrChapterDocInvalidData へ翻訳する（400 として返すため）。
func mapChapterDocError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "22") {
		return repository.ErrChapterDocInvalidData
	}
	return err
}

func (r *teachingMaterialRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil // 存在し得ない id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).DeleteChapter(ctx, id64)
}

// DeleteByCourse はコース削除時の cascade 用に配下教材を全削除する（FK に頼らず明示削除）。
func (r *teachingMaterialRepository) DeleteByCourse(ctx context.Context, courseID uint64) error {
	cid, ok := toInt64ID(courseID)
	if !ok {
		return nil // 存在し得ない course_id = 対象なし
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).DeleteChaptersByCourse(ctx, cid)
}
