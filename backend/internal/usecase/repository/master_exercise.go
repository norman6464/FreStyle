package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// MasterExerciseRepository は運営マスタ演習問題（旧 php_exercises 由来）の永続化を担う。
//
// 旧 PhpExerciseRepository を language 引数を受け取れる形に汎用化したもの。
// 言語フィルタは `ListByLanguage(language string)` で行い、PHP 専用 API は
// usecase 層で `language="php"` を渡して既存挙動と互換を保つ。
//
// 全 メソッド は I/O 境界 として `ctx context.Context` を 第 1 引数 で 受ける
// (キャンセル / タイムアウト / トレース ID 伝搬 の ため)。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// masterExerciseRepository (GORM)。
type MasterExerciseRepository interface {
	ListByLanguage(ctx context.Context, language string) ([]domain.MasterExercise, error)
	GetByID(ctx context.Context, id uint64) (*domain.MasterExercise, error)
	GetBySlug(ctx context.Context, slug string) (*domain.MasterExercise, error)
}
