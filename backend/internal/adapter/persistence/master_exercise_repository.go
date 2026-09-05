package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// masterExerciseRepository は [repository.MasterExerciseRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type masterExerciseRepository struct {
	baseRepository
}

func NewMasterExerciseRepository(db *sql.DB) repository.MasterExerciseRepository {
	return &masterExerciseRepository{baseRepository{db: db}}
}

// nullableChapterID は NULL 可の chapter_id を domain の *uint64 へ写す（未紐付けは nil）。
func nullableChapterID(v sql.NullInt64) *uint64 {
	if !v.Valid {
		return nil
	}
	id := uint64(v.Int64)
	return &id
}

// toDomainMasterExercise は sqlc 生成モデル → domain への詰め替え。
// hint_text / expected_output は NULL 可の列で domain は plain string。NULL は空文字に倒す
// （GORM が NULL を string へ Scan したときと同じ）。
func toDomainMasterExercise(row sqlcgen.MasterExercise) domain.MasterExercise {
	return domain.MasterExercise{
		ID:             uint64(row.ID),
		Slug:           row.Slug,
		Language:       row.Language,
		SortOrder:      int(row.SortOrder),
		Category:       row.Category,
		Title:          row.Title,
		Description:    row.Description,
		StarterCode:    row.StarterCode,
		HintText:       row.HintText.String,
		ExpectedOutput: row.ExpectedOutput.String,
		Mode:           row.Mode,
		Explanation:    row.Explanation,
		Difficulty:     row.Difficulty,
		IsPublished:    row.IsPublished,
		ChapterID:      nullableChapterID(row.ChapterID),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

// ListByLanguage は公開済み問題を返す。language が空文字なら全言語。
func (r *masterExerciseRepository) ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error) {
	rows, err := sqlcgen.New(r.dbtx(ctx)).ListMasterExercisesByLanguage(ctx, language)
	if err != nil {
		return nil, err
	}
	exercises := make([]domain.MasterExercise, 0, len(rows))
	for _, row := range rows {
		exercises = append(exercises, toDomainMasterExercise(row))
	}
	return exercises, nil
}

// GetByID は単一問題を返す。未存在は domain.ErrNotFound（handler が 404 に分岐）。
func (r *masterExerciseRepository) GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, domain.ErrNotFound // 存在し得ない id = not found
	}
	row, err := sqlcgen.New(r.dbtx(ctx)).GetMasterExerciseByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	e := toDomainMasterExercise(row)
	return &e, nil
}

// GetBySlug は slug で単一問題を返す。未存在は domain.ErrNotFound（handler が 404 に分岐）。
func (r *masterExerciseRepository) GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error) {
	row, err := sqlcgen.New(r.dbtx(ctx)).GetMasterExerciseBySlug(ctx, slug)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound // 404 シグナルを維持
	}
	if err != nil {
		return nil, err
	}
	e := toDomainMasterExercise(row)
	return &e, nil
}

// SummaryByLanguage は公開済み問題を言語ごとに集計し、問題数と current user の正解済み件数を返す。
// 言語選択カード用。問題本文を返さないので一覧 API より軽い。
// userID=0（未ログイン）は usr サブクエリが空になり solved は全て 0。
func (r *masterExerciseRepository) SummaryByLanguage(ctx context.Context, userID uint64) ([]repository.ExerciseLanguageSummary, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		uid = 0 // 存在し得ない user_id = どの提出にも一致しない（未ログインと同じ集計）
	}
	rows, err := sqlcgen.New(r.dbtx(ctx)).SummarizeMasterExercisesByLanguage(ctx, sqlcgen.SummarizeMasterExercisesByLanguageParams{
		ExerciseKind: domain.ExerciseKindMaster,
		ViewerID:     uid,
	})
	if err != nil {
		return nil, err
	}
	out := make([]repository.ExerciseLanguageSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.ExerciseLanguageSummary{
			Language: row.Language,
			Total:    row.Total,
			Solved:   row.Solved,
		})
	}
	return out, nil
}

// ListWithStatusByLanguage は公開済み問題を「current user の提出状態 + 全体集計」付きで 1 クエリで返す。
// in.Limit > 0 のとき LIMIT/OFFSET を適用する。Limit=0 は全件取得。
// userID=0（未ログイン）は usr サブクエリが空になり status は全て "".
// sort_order は一意でないため id をタイブレークに置く。これが無いと同値行の順序がページ間で
// 揺れ、OFFSET ページングで同じ行が重複したり抜け落ちたりする。
func (r *masterExerciseRepository) ListWithStatusByLanguage(ctx context.Context, in repository.ListWithStatusInput) ([]repository.MasterExerciseWithStatus, error) {
	uid, ok := toInt64ID(in.UserID)
	if !ok {
		uid = 0 // 存在し得ない user_id = どの提出にも一致しない（未ログインと同じ状態）
	}
	// Limit<=0 は「全件」。SQL 側は 0 を NULL（無制限）へ倒すので 0 に正規化する。
	// 負値をそのまま渡すと LIMIT -1 で SQL エラーになるためここで潰す。
	// Offset は Limit が無いときには意味を持たない（GORM 版も Limit>0 のときだけ適用していた）。
	limit, offset := int64(0), int64(0)
	if in.Limit > 0 {
		limit = int64(in.Limit)
		if in.Offset > 0 {
			offset = int64(in.Offset)
		}
	}
	rows, err := sqlcgen.New(r.dbtx(ctx)).ListMasterExercisesWithStatusByLanguage(ctx, sqlcgen.ListMasterExercisesWithStatusByLanguageParams{
		ExerciseKind: domain.ExerciseKindMaster,
		ViewerID:     uid,
		Language:     in.Language,
		RowLimit:     limit,
		RowOffset:    offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]repository.MasterExerciseWithStatus, 0, len(rows))
	for _, row := range rows {
		out = append(out, repository.MasterExerciseWithStatus{
			MasterExercise: toDomainMasterExercise(sqlcgen.MasterExercise{
				ID:             row.ID,
				Slug:           row.Slug,
				Language:       row.Language,
				SortOrder:      row.SortOrder,
				Category:       row.Category,
				Title:          row.Title,
				Description:    row.Description,
				StarterCode:    row.StarterCode,
				HintText:       row.HintText,
				ExpectedOutput: row.ExpectedOutput,
				Mode:           row.Mode,
				Explanation:    row.Explanation,
				Difficulty:     row.Difficulty,
				IsPublished:    row.IsPublished,
				ChapterID:      row.ChapterID,
				CreatedAt:      row.CreatedAt,
				UpdatedAt:      row.UpdatedAt,
			}),
			Status: row.UserStatus,
			Stats: repository.ExerciseSubmissionStats{
				TotalSubmissions: row.TotalSubmissions,
				SolvedUsers:      row.SolvedUsers,
			},
		})
	}
	return out, nil
}
