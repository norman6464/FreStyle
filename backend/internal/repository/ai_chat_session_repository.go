package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// AiChatSessionRepository は ai_chat_sessions テーブルへのアクセスを提供する。
type AiChatSessionRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.AiChatSession, error)
	FindByID(ctx context.Context, id uint64) (*domain.AiChatSession, error)
	Create(ctx context.Context, s *domain.AiChatSession) error
	UpdateTitle(ctx context.Context, id uint64, title string) error
	Delete(ctx context.Context, id uint64) error
}

type aiChatSessionRepository struct{ db *gorm.DB }

func NewAiChatSessionRepository(db *gorm.DB) AiChatSessionRepository {
	return &aiChatSessionRepository{db: db}
}

func (r *aiChatSessionRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.AiChatSession, error) {
	var rows []domain.AiChatSession
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *aiChatSessionRepository) FindByID(ctx context.Context, id uint64) (*domain.AiChatSession, error) {
	var s domain.AiChatSession
	if err := r.db.WithContext(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *aiChatSessionRepository) Create(ctx context.Context, s *domain.AiChatSession) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *aiChatSessionRepository) UpdateTitle(ctx context.Context, id uint64, title string) error {
	return r.db.WithContext(ctx).
		Model(&domain.AiChatSession{}).
		Where("id = ?", id).
		Update("title", title).Error
}

func (r *aiChatSessionRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.AiChatSession{}, id).Error
}
