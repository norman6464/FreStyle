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

// FollowUserUseCase は単方向フォロー（accepted の Friendship を即時作成）。
// フロントの follow ボタンが確認なしで完了する仕様に合わせる。
type FollowUserUseCase struct{ repo repository.FriendshipRepository }

func NewFollowUserUseCase(r repository.FriendshipRepository) *FollowUserUseCase {
	return &FollowUserUseCase{repo: r}
}

func (u *FollowUserUseCase) Execute(ctx context.Context, followerID, targetID uint64) (*domain.Friendship, error) {
	if followerID == 0 || targetID == 0 {
		return nil, errors.New("followerID and targetID are required")
	}
	if followerID == targetID {
		return nil, errors.New("cannot follow self")
	}
	if existing, err := u.repo.FindBetween(ctx, followerID, targetID); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}
	f := &domain.Friendship{
		RequesterID: followerID,
		AddresseeID: targetID,
		Status:      domain.FriendshipStatusAccepted,
	}
	if err := u.repo.Create(ctx, f); err != nil {
		return nil, err
	}
	return f, nil
}

type UnfollowUserUseCase struct{ repo repository.FriendshipRepository }

func NewUnfollowUserUseCase(r repository.FriendshipRepository) *UnfollowUserUseCase {
	return &UnfollowUserUseCase{repo: r}
}

func (u *UnfollowUserUseCase) Execute(ctx context.Context, followerID, targetID uint64) error {
	if followerID == 0 || targetID == 0 {
		return errors.New("followerID and targetID are required")
	}
	return u.repo.DeleteBetween(ctx, followerID, targetID)
}

type ListFollowingUseCase struct{ repo repository.FriendshipRepository }

func NewListFollowingUseCase(r repository.FriendshipRepository) *ListFollowingUseCase {
	return &ListFollowingUseCase{repo: r}
}

func (u *ListFollowingUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListAcceptedFollowing(ctx, userID)
}

type ListFollowersUseCase struct{ repo repository.FriendshipRepository }

func NewListFollowersUseCase(r repository.FriendshipRepository) *ListFollowersUseCase {
	return &ListFollowersUseCase{repo: r}
}

func (u *ListFollowersUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListAcceptedFollowers(ctx, userID)
}

// FollowStatus は viewer→target、target→viewer の両方向 follow 状態を表す。
type FollowStatus struct {
	IsFollowing  bool `json:"isFollowing"`
	IsFollowedBy bool `json:"isFollowedBy"`
}

type GetFollowStatusUseCase struct{ repo repository.FriendshipRepository }

func NewGetFollowStatusUseCase(r repository.FriendshipRepository) *GetFollowStatusUseCase {
	return &GetFollowStatusUseCase{repo: r}
}

func (u *GetFollowStatusUseCase) Execute(ctx context.Context, viewerID, targetID uint64) (FollowStatus, error) {
	if viewerID == 0 || targetID == 0 {
		return FollowStatus{}, errors.New("viewerID and targetID are required")
	}
	out := FollowStatus{}
	if f, err := u.repo.FindBetween(ctx, viewerID, targetID); err != nil {
		return FollowStatus{}, err
	} else if f != nil && f.Status == domain.FriendshipStatusAccepted {
		out.IsFollowing = true
	}
	if f, err := u.repo.FindBetween(ctx, targetID, viewerID); err != nil {
		return FollowStatus{}, err
	} else if f != nil && f.Status == domain.FriendshipStatusAccepted {
		out.IsFollowedBy = true
	}
	return out, nil
}
