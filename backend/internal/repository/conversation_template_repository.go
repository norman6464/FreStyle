package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ConversationTemplateRepository interface {
	ListActive(ctx context.Context) ([]domain.ConversationTemplate, error)
}

type conversationTemplateRepository struct{ db *gorm.DB }

func NewConversationTemplateRepository(db *gorm.DB) ConversationTemplateRepository {
	return &conversationTemplateRepository{db: db}
}

func (r *conversationTemplateRepository) ListActive(ctx context.Context) ([]domain.ConversationTemplate, error) {
	var rows []domain.ConversationTemplate
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("category, id").Find(&rows).Error
	return rows, err
}
