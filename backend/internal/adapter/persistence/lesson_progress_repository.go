package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// lessonProgressRepository は [repository.LessonProgressRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type lessonProgressRepository struct {
	db *gorm.DB
}

func NewLessonProgressRepository(db *gorm.DB) repository.LessonProgressRepository {
	return &lessonProgressRepository{db: db}
}

func toDomainUserLessonProgress(row sqlcgen.UserChapterProgress) domain.UserLessonProgress {
	return domain.UserLessonProgress{
		ID:     uint64(row.ID),
		UserID: uint64(row.UserID),
		// TeachingMaterialID は列 chapter_id に対応する（JSON は互換のため teachingMaterialId）。
		TeachingMaterialID: uint64(row.ChapterID),
		CourseID:           uint64(row.CourseID),
		CompletedAt:        row.CompletedAt,
		CreatedAt:          row.CreatedAt,
	}
}

func (r *lessonProgressRepository) MarkCompleted(ctx context.Context, userID, materialID, courseID uint64) (bool, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return false, nil // 存在し得ない user_id は記録しない
	}
	cid, ok := toInt64ID(materialID)
	if !ok {
		return false, nil
	}
	coid, ok := toInt64ID(courseID)
	if !ok {
		return false, nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return false, err
	}
	// (user_id, chapter_id) が衝突したら何もしない（冪等）。RowsAffected>0 で初回かを判定する。
	n, err := sqlcgen.New(sqlDB).InsertUserChapterProgressIfAbsent(ctx, sqlcgen.InsertUserChapterProgressIfAbsentParams{
		UserID:    uid,
		ChapterID: cid,
		CourseID:  coid,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *lessonProgressRepository) MarkIncomplete(ctx context.Context, userID, materialID uint64) error {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil
	}
	cid, ok := toInt64ID(materialID)
	if !ok {
		return nil
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlcgen.New(sqlDB).DeleteUserChapterProgress(ctx, sqlcgen.DeleteUserChapterProgressParams{
		UserID:    uid,
		ChapterID: cid,
	})
}

// CountCompletedByUserGroupedByCourse は「現存する published 教材」の完了行のみを
// course_id ごとに 1 クエリで集計する。教材削除で JOIN から落ち、非公開化は is_published で
// 除外されるため、分子が分母(published 章数)を上回ることはない。
func (r *lessonProgressRepository) CountCompletedByUserGroupedByCourse(ctx context.Context, userID uint64) (map[uint64]int, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return map[uint64]int{}, nil // 存在し得ない user_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).CountCompletedChaptersByCourseForUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	counts := make(map[uint64]int, len(rows))
	for _, row := range rows {
		counts[uint64(row.CourseID)] = int(row.Cnt)
	}
	return counts, nil
}

func (r *lessonProgressRepository) ListByUser(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error) {
	uid, ok := toInt64ID(userID)
	if !ok {
		return []domain.UserLessonProgress{}, nil // 存在し得ない user_id = 0 件
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListUserChapterProgressByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.UserLessonProgress, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainUserLessonProgress(row))
	}
	return out, nil
}
