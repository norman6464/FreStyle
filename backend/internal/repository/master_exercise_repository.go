package repository

import (
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// MasterExerciseRepository は運営マスタ演習問題（旧 php_exercises 由来）の永続化を担う。
//
// 旧 PhpExerciseRepository を language 引数を受け取れる形に汎用化したもの。
// 言語フィルタは `ListByLanguage(language string)` で行い、PHP 専用 API は
// usecase 層で `language="php"` を渡して既存挙動と互換を保つ。
type MasterExerciseRepository interface {
	ListByLanguage(language string) ([]domain.MasterExercise, error)
	GetByID(id uint64) (*domain.MasterExercise, error)
	GetBySlug(slug string) (*domain.MasterExercise, error)
}

type masterExerciseRepository struct {
	db *gorm.DB
}

func NewMasterExerciseRepository(db *gorm.DB) MasterExerciseRepository {
	return &masterExerciseRepository{db: db}
}

func (r *masterExerciseRepository) ListByLanguage(language string) ([]domain.MasterExercise, error) {
	var exercises []domain.MasterExercise
	q := r.db.Where("is_published = ?", true)
	if language != "" {
		q = q.Where("language = ?", language)
	}
	if err := q.Order("order_index asc").Find(&exercises).Error; err != nil {
		return nil, err
	}
	return exercises, nil
}

func (r *masterExerciseRepository) GetByID(id uint64) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.First(&exercise, id).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}

func (r *masterExerciseRepository) GetBySlug(slug string) (*domain.MasterExercise, error) {
	var exercise domain.MasterExercise
	if err := r.db.Where("slug = ?", slug).First(&exercise).Error; err != nil {
		return nil, err
	}
	return &exercise, nil
}
