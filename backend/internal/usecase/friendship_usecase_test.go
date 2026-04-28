package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubFriendshipRepo struct {
	rows      []domain.Friendship
	err       error
	between   *domain.Friendship
	deleteHit struct{ requesterID, addresseeID uint64 }
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
func (s *stubFriendshipRepo) ListAcceptedFollowing(_ context.Context, _ uint64) ([]domain.Friendship, error) {
	return s.rows, s.err
}
func (s *stubFriendshipRepo) ListAcceptedFollowers(_ context.Context, _ uint64) ([]domain.Friendship, error) {
	return s.rows, s.err
}
func (s *stubFriendshipRepo) FindBetween(_ context.Context, _, _ uint64) (*domain.Friendship, error) {
	return s.between, s.err
}
func (s *stubFriendshipRepo) DeleteBetween(_ context.Context, requesterID, addresseeID uint64) error {
	s.deleteHit.requesterID = requesterID
	s.deleteHit.addresseeID = addresseeID
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

func TestFollowUser_RejectsSelf(t *testing.T) {
	uc := NewFollowUserUseCase(&stubFriendshipRepo{})
	if _, err := uc.Execute(context.Background(), 1, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestFollowUser_ReturnsExistingIfAlreadyFollowing(t *testing.T) {
	existing := &domain.Friendship{ID: 99, RequesterID: 1, AddresseeID: 2, Status: domain.FriendshipStatusAccepted}
	uc := NewFollowUserUseCase(&stubFriendshipRepo{between: existing})
	got, err := uc.Execute(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ID != 99 {
		t.Fatalf("should return existing, got %+v", got)
	}
}

func TestFollowUser_CreatesAcceptedFriendship(t *testing.T) {
	uc := NewFollowUserUseCase(&stubFriendshipRepo{})
	got, err := uc.Execute(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.Status != domain.FriendshipStatusAccepted {
		t.Fatalf("status should be accepted, got %s", got.Status)
	}
}

func TestUnfollowUser_PassesIDsToRepo(t *testing.T) {
	repo := &stubFriendshipRepo{}
	uc := NewUnfollowUserUseCase(repo)
	if err := uc.Execute(context.Background(), 7, 11); err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.deleteHit.requesterID != 7 || repo.deleteHit.addresseeID != 11 {
		t.Fatalf("Delete called with wrong ids: %+v", repo.deleteHit)
	}
}

func TestGetFollowStatus_RequiresIDs(t *testing.T) {
	uc := NewGetFollowStatusUseCase(&stubFriendshipRepo{})
	if _, err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error")
	}
}
