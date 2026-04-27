package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type NoteRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error)
	FindByID(ctx context.Context, id uint64) (*domain.Note, error)
	Create(ctx context.Context, n *domain.Note) error
	Update(ctx context.Context, n *domain.Note) error
	Delete(ctx context.Context, id uint64) error
}

type noteRepository struct{ db *gorm.DB }

func NewNoteRepository(db *gorm.DB) NoteRepository { return &noteRepository{db: db} }

func (r *noteRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Note, error) {
	var rows []domain.Note
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC").Find(&rows).Error
	return rows, err
}

func (r *noteRepository) FindByID(ctx context.Context, id uint64) (*domain.Note, error) {
	var n domain.Note
	if err := r.db.WithContext(ctx).First(&n, id).Error; err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *noteRepository) Create(ctx context.Context, n *domain.Note) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *noteRepository) Update(ctx context.Context, n *domain.Note) error {
	return r.db.WithContext(ctx).Save(n).Error
}

func (r *noteRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Note{}, id).Error
}
