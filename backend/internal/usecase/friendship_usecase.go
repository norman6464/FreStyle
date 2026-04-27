package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListFriendshipsUseCase struct{ repo repository.FriendshipRepository }

func NewListFriendshipsUseCase(r repository.FriendshipRepository) *ListFriendshipsUseCase {
	return &ListFriendshipsUseCase{repo: r}
}

func (u *ListFriendshipsUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type RequestFriendshipUseCase struct{ repo repository.FriendshipRepository }

func NewRequestFriendshipUseCase(r repository.FriendshipRepository) *RequestFriendshipUseCase {
	return &RequestFriendshipUseCase{repo: r}
}

func (u *RequestFriendshipUseCase) Execute(ctx context.Context, requesterID, addresseeID uint64) (*domain.Friendship, error) {
	if requesterID == 0 || addresseeID == 0 {
		return nil, errors.New("requesterID and addresseeID are required")
	}
	if requesterID == addresseeID {
		return nil, errors.New("cannot request friendship to self")
	}
	f := &domain.Friendship{RequesterID: requesterID, AddresseeID: addresseeID, Status: domain.FriendshipStatusPending}
	if err := u.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

type RespondFriendshipUseCase struct{ repo repository.FriendshipRepository }

func NewRespondFriendshipUseCase(r repository.FriendshipRepository) *RespondFriendshipUseCase {
	return &RespondFriendshipUseCase{repo: r}
}

func (u *RespondFriendshipUseCase) Execute(ctx context.Context, id uint64, accept bool) error {
	if id == 0 {
		return errors.New("id is required")
	}
	status := domain.FriendshipStatusRejected
	if accept {
		status = domain.FriendshipStatusAccepted
	}
	return u.repo.UpdateStatus(ctx, id, status)
}
