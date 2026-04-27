package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubChatRoomRepo struct {
	rows []domain.ChatRoom
	err  error
}

func (s *stubChatRoomRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.ChatRoom, error) {
	return s.rows, s.err
}
func (s *stubChatRoomRepo) Create(_ context.Context, r *domain.ChatRoom) error {
	if s.err != nil {
		return s.err
	}
	r.ID = 99
	return nil
}

func TestGetChatRoomsByUserID_RequiresUserID(t *testing.T) {
	uc := NewGetChatRoomsByUserIDUseCase(&stubChatRoomRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetChatRoomsByUserID_Returns(t *testing.T) {
	uc := NewGetChatRoomsByUserIDUseCase(&stubChatRoomRepo{rows: []domain.ChatRoom{{ID: 1}}})
	got, err := uc.Execute(context.Background(), 5)
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: rows=%v err=%v", got, err)
	}
}

func TestCreateChatRoom_RequiresName(t *testing.T) {
	uc := NewCreateChatRoomUseCase(&stubChatRoomRepo{})
	if _, err := uc.Execute(context.Background(), "", false); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateChatRoom_AssignsID(t *testing.T) {
	uc := NewCreateChatRoomUseCase(&stubChatRoomRepo{})
	got, err := uc.Execute(context.Background(), "team-a", true)
	if err != nil || got.ID != 99 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
