package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// lessonProgressRepository は [repository.LessonProgressRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、接続プール（*sql.DB）をそのまま受け取る。
type lessonProgressRepository struct {
	db *sql.DB
}

func NewLessonProgressRepository(db *sql.DB) repository.LessonProgressRepository {
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
	// (user_id, chapter_id) が衝突したら何もしない（冪等）。RowsAffected>0 で初回かを判定する。
	n, err := sqlcgen.New(r.db).InsertUserChapterProgressIfAbsent(ctx, sqlcgen.InsertUserChapterProgressIfAbsentParams{
		UserID:    uid,
		ChapterID: cid,
		CourseID:  coid,
	})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// MarkIncomplete は章の完了記録を取り消す（行を消す）。
//
// ここは 0 行削除を成功のままにする（not-found にしない）。この操作が表しているのは
// 「その (user, chapter) の完了記録が無い状態」で、まだ完了していない章のチェックを外す
// 要求は最初からその事後条件を満たしている。MarkCompleted 側も ON CONFLICT DO NOTHING の
// 冪等な INSERT なので、完了トグルは往復のどちらでも「今の状態に揃える」意味になる。
// 未完了の章を外しただけで 404 が返ると、フロントの楽観更新がロールバックされて
// チェックが勝手に戻る（利用者から見れば操作が拒否されたのと区別が付かない）。
func (r *lessonProgressRepository) MarkIncomplete(ctx context.Context, userID, materialID uint64) error {
	uid, ok := toInt64ID(userID)
	if !ok {
		return nil // 存在し得ない user_id = 消す記録がない（冪等に成功）
	}
	cid, ok := toInt64ID(materialID)
	if !ok {
		return nil // 存在し得ない chapter_id = 消す記録がない（冪等に成功）
	}
	return sqlcgen.New(r.db).DeleteUserChapterProgress(ctx, sqlcgen.DeleteUserChapterProgressParams{
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
	rows, err := sqlcgen.New(r.db).CountCompletedChaptersByCourseForUser(ctx, uid)
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
	rows, err := sqlcgen.New(r.db).ListUserChapterProgressByUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.UserLessonProgress, 0, len(rows))
	for _, row := range rows {
		out = append(out, toDomainUserLessonProgress(row))
	}
	return out, nil
}
