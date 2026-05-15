package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AiChatSessionRepository は ai_chat_sessions テーブルへのアクセスを提供する。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// aiChatSessionRepository (GORM)。
type AiChatSessionRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.AiChatSession, error)
	FindByID(ctx context.Context, id uint64) (*domain.AiChatSession, error)
	Create(ctx context.Context, s *domain.AiChatSession) error
	UpdateTitle(ctx context.Context, id uint64, title string) error
	Delete(ctx context.Context, id uint64) error
}
