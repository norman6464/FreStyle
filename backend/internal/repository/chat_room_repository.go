package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ChatRoomRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.ChatRoom, error)
	Create(ctx context.Context, r *domain.ChatRoom) error
}

type chatRoomRepository struct{ db *gorm.DB }

func NewChatRoomRepository(db *gorm.DB) ChatRoomRepository {
	return &chatRoomRepository{db: db}
}

func (r *chatRoomRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.ChatRoom, error) {
	var rows []domain.ChatRoom
	err := r.db.WithContext(ctx).
		Joins("JOIN chat_room_members m ON m.room_id = chat_rooms.id").
		Where("m.user_id = ?", userID).
		Order("chat_rooms.created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *chatRoomRepository) Create(ctx context.Context, room *domain.ChatRoom) error {
	return r.db.WithContext(ctx).Create(room).Error
}
