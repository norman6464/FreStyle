package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type SessionNoteRepository interface {
	FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error)
	Upsert(ctx context.Context, n *domain.SessionNote) error
}

type sessionNoteRepository struct{ db *gorm.DB }

func NewSessionNoteRepository(db *gorm.DB) SessionNoteRepository {
	return &sessionNoteRepository{db: db}
}

func (r *sessionNoteRepository) FindBySessionID(ctx context.Context, sessionID uint64) (*domain.SessionNote, error) {
	var n domain.SessionNote
	err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *sessionNoteRepository) Upsert(ctx context.Context, n *domain.SessionNote) error {
	return r.db.WithContext(ctx).Save(n).Error
}
