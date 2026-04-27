package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetChatRoomsByUserIDUseCase struct {
	rooms repository.ChatRoomRepository
}

func NewGetChatRoomsByUserIDUseCase(r repository.ChatRoomRepository) *GetChatRoomsByUserIDUseCase {
	return &GetChatRoomsByUserIDUseCase{rooms: r}
}

func (u *GetChatRoomsByUserIDUseCase) Execute(ctx context.Context, userID uint64) ([]domain.ChatRoom, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.rooms.ListByUserID(ctx, userID)
}

type CreateChatRoomUseCase struct {
	rooms repository.ChatRoomRepository
}

func NewCreateChatRoomUseCase(r repository.ChatRoomRepository) *CreateChatRoomUseCase {
	return &CreateChatRoomUseCase{rooms: r}
}

func (u *CreateChatRoomUseCase) Execute(ctx context.Context, name string, isGroup bool) (*domain.ChatRoom, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	r := &domain.ChatRoom{Name: name, IsGroup: isGroup}
	if err := u.rooms.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}
