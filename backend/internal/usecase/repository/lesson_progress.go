package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// LessonProgressRepository は教材（レッスン）の完了状態を永続化する。
type LessonProgressRepository interface {
	// MarkCompleted は (user, material) を完了として記録する（既に完了済みなら冪等）。
	// 戻り値の bool は初回完了時のみ true。重複リクエストでは false を返す。
	MarkCompleted(ctx context.Context, userID, materialID, courseID uint64) (bool, error)
	// MarkIncomplete は完了記録を取り消す（行削除。未記録でもエラーにしない）。
	MarkIncomplete(ctx context.Context, userID, materialID uint64) error
	// ListByUser は user の完了記録をすべて返す。
	ListByUser(ctx context.Context, userID uint64) ([]domain.UserLessonProgress, error)
	// CountCompletedByUserGroupedByCourse は user の完了章数を course_id ごとに一括集計して返す。
	// 現存する published 教材への JOIN で数える(非公開化・削除された章の完了行は含めない =
	// コース詳細ページの分子と同じ意味論。分母 CountByCourseForCompany の published 側と揃う)。
	CountCompletedByUserGroupedByCourse(ctx context.Context, userID uint64) (map[uint64]int, error)
}
