package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubFriendshipRepo struct {
	rows []domain.Friendship
	err  error
}

func (s *stubFriendshipRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.Friendship, error) {
	return s.rows, s.err
}
func (s *stubFriendshipRepo) Create(_ context.Context, f *domain.Friendship) error {
	if s.err != nil {
		return s.err
	}
	f.ID = 81
	return nil
}
func (s *stubFriendshipRepo) UpdateStatus(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func TestListFriendships_RequiresUserID(t *testing.T) {
	uc := NewListFriendshipsUseCase(&stubFriendshipRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestRequestFriendship_RejectsSelf(t *testing.T) {
	uc := NewRequestFriendshipUseCase(&stubFriendshipRepo{})
	if _, err := uc.Execute(context.Background(), 1, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestRequestFriendship_AssignsID(t *testing.T) {
	uc := NewRequestFriendshipUseCase(&stubFriendshipRepo{})
	got, err := uc.Execute(context.Background(), 1, 2)
	if err != nil || got.ID != 81 || got.Status != domain.FriendshipStatusPending {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestRespondFriendship_RequiresID(t *testing.T) {
	uc := NewRespondFriendshipUseCase(&stubFriendshipRepo{})
	if err := uc.Execute(context.Background(), 0, true); err == nil {
		t.Fatal("expected error")
	}
}
