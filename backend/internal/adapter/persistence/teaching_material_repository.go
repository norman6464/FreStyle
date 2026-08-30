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
)

// teachingMaterialRepository は [repository.TeachingMaterialRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type teachingMaterialRepository struct {
	db *sql.DB
}

func NewTeachingMaterialRepository(db *sql.DB) repository.TeachingMaterialRepository {
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

// toDomainChapter は行全体（本文 doc・workspace_id 含む）を domain へ写す。
// chapterRow は行全体を返すクエリの共通の行形。GetChapterByID / UpdateChapterDocWithRevision は
// course_chapters の全列を返すため、sqlc は個別の Row 型を生成せずテーブル型
// （sqlcgen.CourseChapter）をそのまま再利用する（FRESTYLE-403 で workspace_id を追加した際、
// 既存の列リストと一致し全列になったことでこの型に切り替わった）。
type chapterRow = sqlcgen.CourseChapter

func toDomainChapter(row chapterRow) domain.TeachingMaterial {
	m := domain.TeachingMaterial{
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
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		m.WorkspaceID = &wid
	}
	return m
}

// toDomainChapterSummary は一覧用の軽量行（本文 doc を含まない）を domain へ写す。Doc は nil のまま。
//
// sqlc は SELECT ごとに別の行型を生成するため、company 別の一覧は
// [toDomainChapterCompanySummary] が同じ列構成であることを型変換で確かめてからここへ渡す。
func toDomainChapterSummary(row sqlcgen.ListChaptersByCourseRow) domain.TeachingMaterial {
	m := domain.TeachingMaterial{
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
	if row.WorkspaceID.Valid {
		wid := row.WorkspaceID.UUID.String()
		m.WorkspaceID = &wid
	}
	return m
}

// toDomainChapterCompanySummary は company 別一覧の軽量行を domain へ写す。
// 列構成が course 別一覧とずれたらこの型変換がコンパイルエラーになる。
func toDomainChapterCompanySummary(row sqlcgen.ListChaptersByCompanyRow) domain.TeachingMaterial {
	return toDomainChapterSummary(sqlcgen.ListChaptersByCourseRow(row))
}

// toDomainChapterWorkspaceSummary は workspace 別一覧の軽量行を domain へ写す。
// 列構成が course 別一覧とずれたらこの型変換がコンパイルエラーになる。
func toDomainChapterWorkspaceSummary(row sqlcgen.ListChaptersByWorkspaceRow) domain.TeachingMaterial {
	return toDomainChapterSummary(sqlcgen.ListChaptersByCourseRow(row))
}

// ListByCompany は backward-compat 用（コース対応完了後に削除予定）。
func (r *teachingMaterialRepository) ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	cid, ok := toInt64ID(companyID)
	if !ok {
		return []domain.TeachingMaterial{}, nil // 存在し得ない company_id = 0 件
	}
	// 一覧は本文（doc・jsonb）を返さない（domain の Doc は json:"-" で応答に出ないため、
	// 全章分を読み出しても転送するだけ無駄になる）。ListByCourse と同じ軽量な列構成。
	rows, err := sqlcgen.New(r.db).ListChaptersByCompany(ctx, sqlcgen.ListChaptersByCompanyParams{
		CompanyID:          cid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	materials := make([]domain.TeachingMaterial, 0, len(rows))
	for _, row := range rows {
		materials = append(materials, toDomainChapterCompanySummary(row))
	}
	return materials, nil
}

// ListByWorkspace は backward-compat 用（コース対応完了後に削除予定）。ListByCompany の
// workspace_id 版で、TeachingMaterialUseCase.List が使う現行の経路。
func (r *teachingMaterialRepository) ListByWorkspace(ctx context.Context, workspaceID string, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return []domain.TeachingMaterial{}, nil // 不正 / 空の ID は該当なしと同じ扱い
	}
	// 一覧は本文（doc・jsonb）を返さない（ListByCompany と同じ軽量な列構成）。
	rows, err := sqlcgen.New(r.db).ListChaptersByWorkspace(ctx, sqlcgen.ListChaptersByWorkspaceParams{
		WorkspaceID:        wid,
		IncludeUnpublished: includeUnpublished,
	})
	if err != nil {
		return nil, err
	}
	materials := make([]domain.TeachingMaterial, 0, len(rows))
	for _, row := range rows {
		materials = append(materials, toDomainChapterWorkspaceSummary(row))
	}
	return materials, nil
}

// ListByCourse はコース内の章を sort_order 昇順で返す。
func (r *teachingMaterialRepository) ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error) {
	cid, ok := toInt64ID(courseID)
	if !ok {
		return []domain.TeachingMaterial{}, nil // 存在し得ない course_id = 0 件
	}
	// 一覧は本文（doc・jsonb）を返さない（章ごとに重く、全章を先読みすると非効率）。
	// 本文は選択時に GetByID で都度取得する。Doc は nil のままになる。
	rows, err := sqlcgen.New(r.db).ListChaptersByCourse(ctx, sqlcgen.ListChaptersByCourseParams{
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

// GetByID は単一教材を返す。未存在は domain.ErrNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	row, err := sqlcgen.New(r.db).GetChapterByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	m := toDomainChapter(row)
	return &m, nil
}

// CountByCourseForCompany は course_id ごとの教材件数を 1 クエリで集計する。
// trainee 向け(includeUnpublished=false)は published のみ数え、コース詳細の進捗分母と一致させる。
func (r *teachingMaterialRepository) CountByCourseForWorkspace(ctx context.Context, workspaceID string, includeUnpublished bool) (map[uint64]int, error) {
	wid, ok := toNullUUID(workspaceID)
	if !ok {
		return map[uint64]int{}, nil // 不正 / 空の ID は該当なしと同じ扱い
	}
	rows, err := sqlcgen.New(r.db).CountChaptersByCourseForWorkspace(ctx, sqlcgen.CountChaptersByCourseForWorkspaceParams{
		WorkspaceID:        wid,
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
	now := time.Now()
	createdAt := m.CreatedAt
	if createdAt.IsZero() {
		createdAt = now // GORM autoCreateTime 相当（ゼロのときだけ now）
	}
	updatedAt := m.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now // GORM autoUpdateTime 相当（ゼロのときだけ now）
	}
	row, err := sqlcgen.New(r.db).InsertChapter(ctx, sqlcgen.InsertChapterParams{
		CompanyID:       companyID,
		CourseID:        courseID,
		CreatedByUserID: createdBy,
		Title:           m.Title,
		Revision:        int64(m.Revision),      // 0 は SQL 側の COALESCE で既定 1 に倒す
		SchemaVersion:   int64(m.SchemaVersion), // 0 は SQL 側の COALESCE で既定 1 に倒す
		SortOrder:       int64(m.OrderInCourse), // 0 は SQL 側の COALESCE で既定 100 に倒す
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
// domain.ErrNotFound（handler が 404 に分岐）。
func (r *teachingMaterialRepository) Update(ctx context.Context, m *domain.TeachingMaterial) error {
	id64, ok := toInt64ID(m.ID)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = not found
	}
	// CreatedBy / CompanyID / CourseID / Doc / Revision は不変（GORM の Updates(map) と同じ 3 列のみ）。
	// updated_at は now() へ進めて RETURNING で書き戻す（autoUpdateTime 相当）。
	updatedAt, err := sqlcgen.New(r.db).UpdateChapter(ctx, sqlcgen.UpdateChapterParams{
		ID:          id64,
		Title:       m.Title,
		SortOrder:   int64(m.OrderInCourse),
		IsPublished: m.IsPublished,
	})
	if err != nil {
		// 0 行 = 取得と更新の間に章が消えた。黙って nil を返すと失われた編集を保存済みに見せるので 404 を返す。
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
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
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	raw := json.RawMessage(doc)
	row, err := sqlcgen.New(r.db).UpdateChapterDocWithRevision(ctx, sqlcgen.UpdateChapterDocWithRevisionParams{
		ID:               id64,
		Doc:              &raw,
		ExpectedRevision: int64(expectedRevision),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// 0 行 = 「存在しない」か「版不一致」。存在確認で切り分ける。
			if _, gerr := r.GetByID(ctx, id); gerr != nil {
				return nil, gerr // domain.ErrNotFound
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

// Delete は教材を物理削除する。対象行が無ければ domain.ErrNotFound を返す。
//
// DELETE でも 0 行を成功にしない理由:
//
//	「消えている」という事後条件だけなら 0 行削除も満たしている。それでも not-found を
//	返すのは、この経路が「管理者が一覧から 1 件選んで消す」操作だから。0 行を 204 で
//	返すと、既に他の管理者が消した行・別会社の行を消したつもりになり、消えていないものを
//	消したと誤認したまま画面から消える。
//	usecase（TeachingMaterialUseCase.Delete）は GetByID で存在と会社を先に確かめるので、
//	ここに落ちるのは「確認と削除のあいだに章が消えた」競合のときだけ。
//	同じ判断を UpdateMeta 側（0 行 = sql.ErrNoRows を domain.ErrNotFound へ）で既に取っており、
//	更新と削除で結末を揃える。
func (r *teachingMaterialRepository) Delete(ctx context.Context, id uint64) error {
	id64, ok := toInt64ID(id)
	if !ok {
		return domain.ErrNotFound // 存在し得ない id = 対象なし
	}
	// :execrows なので実際に消えた行数が返る（:exec だと 0 行でも成功と区別が付かない）。
	affected, err := sqlcgen.New(r.db).DeleteChapter(ctx, id64)
	if err != nil {
		return err
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteByCourse はコース削除時の cascade 用に配下教材を全削除する（FK に頼らず明示削除）。
//
// ここだけは 0 行を not-found にしない（件数を見ない）。単一行を狙う Delete と違い、
// これは「course_id にぶら下がる行を全部消す」一括操作で、0 行は「教材が 1 つも無い
// コースだった」という正常な結果でしかない。not-found にすると空のコースが削除できなくなる。
func (r *teachingMaterialRepository) DeleteByCourse(ctx context.Context, courseID uint64) error {
	cid, ok := toInt64ID(courseID)
	if !ok {
		return nil // 存在し得ない course_id = 消す相手がいない（一括削除なので 0 件で正常）
	}
	return sqlcgen.New(r.db).DeleteChaptersByCourse(ctx, cid)
}
