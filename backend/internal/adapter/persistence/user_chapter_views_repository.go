package persistence

import (
	"context"
	"database/sql"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userChapterViewRepository は [repository.UserChapterViewRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type userChapterViewRepository struct {
	db *gorm.DB
}

// NewUserChapterViewRepository は UserChapterViewRepository の実装を返す。
func NewUserChapterViewRepository(db *gorm.DB) repository.UserChapterViewRepository {
	return &userChapterViewRepository{db: db}
}

func toDomainUserChapterView(row sqlcgen.UserChapterView) domain.UserChapterView {
	return domain.UserChapterView{
		UserID: uint64(row.UserID),
		// TeachingMaterialID は列 chapter_id に対応する（JSON は互換のため teachingMaterialId）。
		TeachingMaterialID: uint64(row.ChapterID),
		CourseID:           uint64(row.CourseID),
		FirstViewedAt:      row.FirstViewedAt,
		LastViewedAt:       row.LastViewedAt,
		ViewCount:          int(row.ViewCount),
	}
}

// UpsertView は章閲覧を記録する。初回 INSERT、2 回目以降は last_viewed_at と view_count を更新する。
func (r *userChapterViewRepository) UpsertView(
	ctx context.Context,
	userID, teachingMaterialID, courseID uint64,
) error {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil // 存在し得ない user_id は書き込まない
	}
	cid, ok := toInt64ID(teachingMaterialID)
	if !ok {
		return nil
	}
	coid, ok := toInt64ID(courseID)
	if !ok {
		return nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).UpsertUserChapterView(ctx, sqlcgen.UpsertUserChapterViewParams{
		UserID:    uid,
		ChapterID: cid,
		CourseID:  coid,
	})
}

func (r *userChapterViewRepository) ListRecentByUser(
	ctx context.Context,
	userID uint64,
	limit int,
) ([]domain.UserChapterView, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.UserChapterView{}, nil // 存在し得ない user_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListRecentUserChapterViewsByUser(ctx, sqlcgen.ListRecentUserChapterViewsByUserParams{
		UserID:   uid,
		RowLimit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.UserChapterView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainUserChapterView(row))
	}
	return out, nil
}

// GetLastViewedByUserAndCourse は (user, course) の閲覧記録から last_viewed_at 最大の 1 件を返す。
// 履歴なしはエラーではなく (nil, nil)(「初めて開くコース」は正常系のため)。
func (r *userChapterViewRepository) GetLastViewedByUserAndCourse(
	ctx context.Context,
	userID, courseID uint64,
) (*domain.UserChapterView, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil, nil
	}
	coid, ok := toInt64ID(courseID)
	if !ok {
		return nil, nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetLastViewedUserChapterViewByCourse(ctx, sqlcgen.GetLastViewedUserChapterViewByCourseParams{
		UserID:   uid,
		CourseID: coid,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil // 履歴なしは正常系
	}
	if err != nil {
		return nil, err
	}
	v := toDomainUserChapterView(row)
	return &v, nil
}
