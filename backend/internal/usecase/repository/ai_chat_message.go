package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// AiChatMessageRepository は DynamoDB 上の AI チャットメッセージへのアクセスを提供する（PK: sessionId / SK: messageId）。
type AiChatMessageRepository interface {
	Save(ctx context.Context, msg *domain.AiChatMessage) error
	ListBySessionID(ctx context.Context, sessionID uint64) ([]domain.AiChatMessage, error)
}
