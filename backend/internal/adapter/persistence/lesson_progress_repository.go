package persistence

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// lessonProgressRepository は [repository.LessonProgressRepository] の GORM 実装。
type lessonProgressRepository struct {
	db *gorm.DB
}

func NewLessonProgressRepository(db *gorm.DB) repository.LessonProgressRepository {
	return &lessonProgressRepository{db: db}
}

func (r *lessonProgressRepository) MarkCompleted(ctx context.Context, userID, materialID, courseID uint64) (bool, error) {
	completedAt := time.Now()
	row := &domain.UserLessonProgress{
		UserID:             userID,
		TeachingMaterialID: materialID,
		CourseID:           courseID,
		CompletedAt:        completedAt,
	}
	// (user_id, chapter_id) が衝突したら何もしない（冪等）。RowsAffected で初回かを判定。
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "chapter_id"}},
		DoNothing: true,
	}).Create(row)
	if result.Error != nil {
		return false, result.Error
	}
	// Expand フェーズ: 後継テーブル user_chapter_progress にも同じ行を upsert する(dual-write)。
	// Contract で読み書きを新テーブルへ切り替え、旧テーブルを削除する。
	const shadow = `
INSERT INTO user_chapter_progress (user_id, chapter_id, course_id, completed_at, created_at)
VALUES (?, ?, ?, ?, NOW())
ON CONFLICT (user_id, chapter_id) DO NOTHING`
	if err := r.db.WithContext(ctx).Exec(shadow, userID, materialID, courseID, completedAt).Error; err != nil {
		return false, err
	}
	return result.RowsAffected > 0, nil
}

func (r *lessonProgressRepository) MarkIncomplete(ctx context.Context, userID, materialID uint64) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND chapter_id = ?", userID, materialID).
		Delete(&domain.UserLessonProgress{}).Error; err != nil {
		return err
	}
	// Expand フェーズ: 後継テーブルからも削除する(dual-write)。
	return r.db.WithContext(ctx).
		Exec(`DELETE FROM user_chapter_progress WHERE user_id = ? AND chapter_id = ?`, userID, materialID).Error
}

// CountCompletedByUserGroupedByCourse は「現存する published 教材」の完了行のみを
// course_id ごとに 1 クエリで集計する。教材削除で JOIN から落ち、非公開化は is_published で
// 除外されるため、分子が分母(published 章数)を上回ることはない。
func (r *lessonProgressRepository) CountCompletedByUserGroupedByCourse(ctx context.Context, userID uint64) (map[uint64]int, error) {
	const q = `
SELECT tm.course_id, COUNT(*) AS cnt
FROM user_lesson_progress ulp
JOIN course_chapters tm ON tm.id = ulp.chapter_id
WHERE ulp.user_id = ? AND tm.is_published = TRUE
GROUP BY tm.course_id`
	var rows []struct {
		CourseID uint64
		Cnt      int
	}
	if err := r.db.WithContext(ctx).Raw(q, userID).Scan(&rows).Error; err != nil {
		return nil, err
	}
	counts := make(map[uint64]int, len(rows))
	for _, row := range rows {
		counts[row.CourseID] = row.Cnt
	}
	return counts, nil
}

func (r *lessonProgressRepository) ListByUser(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error) {
	var rows []domain.UserLessonProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
