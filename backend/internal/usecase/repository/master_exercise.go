package repository

import "github.com/norman6464/FreStyle/backend/internal/domain"

// MasterExerciseRepository は運営マスタ演習問題（旧 php_exercises 由来）の永続化を担う。
//
// 旧 PhpExerciseRepository を language 引数を受け取れる形に汎用化したもの。
// 言語フィルタは `ListByLanguage(language string)` で行い、PHP 専用 API は
// usecase 層で `language="php"` を渡して既存挙動と互換を保つ。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// masterExerciseRepository (GORM)。
type MasterExerciseRepository interface {
	ListByLanguage(language string) ([]domain.MasterExercise, error)
	GetByID(id uint64) (*domain.MasterExercise, error)
	GetBySlug(slug string) (*domain.MasterExercise, error)
}
