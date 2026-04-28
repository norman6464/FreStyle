package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type FriendshipRepository interface {
	ListByUserID(ctx context.Context, userID uint64) ([]domain.Friendship, error)
	Create(ctx context.Context, f *domain.Friendship) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
	// ListAcceptedFollowing は user が requester かつ accepted な相手（フォロー中）を返す。
	ListAcceptedFollowing(ctx context.Context, userID uint64) ([]domain.Friendship, error)
	// ListAcceptedFollowers は user が addressee かつ accepted な相手（フォロワー）を返す。
	ListAcceptedFollowers(ctx context.Context, userID uint64) ([]domain.Friendship, error)
	// FindBetween は requester→addressee 方向の friendship を 1 件返す（無ければ nil, nil）。
	FindBetween(ctx context.Context, requesterID, addresseeID uint64) (*domain.Friendship, error)
	// DeleteBetween は requester→addressee の friendship を削除する。
	DeleteBetween(ctx context.Context, requesterID, addresseeID uint64) error
}

type friendshipRepository struct{ db *gorm.DB }

func NewFriendshipRepository(db *gorm.DB) FriendshipRepository { return &friendshipRepository{db: db} }

func (r *friendshipRepository) ListByUserID(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	var rows []domain.Friendship
	err := r.db.WithContext(ctx).
		Where("requester_id = ? OR addressee_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *friendshipRepository) Create(ctx context.Context, f *domain.Friendship) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *friendshipRepository) UpdateStatus(ctx context.Context, id uint64, status string) error {
	return r.db.WithContext(ctx).Model(&domain.Friendship{}).Where("id = ?", id).Update("status", status).Error
}

func (r *friendshipRepository) ListAcceptedFollowing(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	var rows []domain.Friendship
	err := r.db.WithContext(ctx).
		Where("requester_id = ? AND status = ?", userID, domain.FriendshipStatusAccepted).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *friendshipRepository) ListAcceptedFollowers(ctx context.Context, userID uint64) ([]domain.Friendship, error) {
	var rows []domain.Friendship
	err := r.db.WithContext(ctx).
		Where("addressee_id = ? AND status = ?", userID, domain.FriendshipStatusAccepted).
		Order("created_at DESC").
		Find(&rows).Error
	return rows, err
}

func (r *friendshipRepository) FindBetween(ctx context.Context, requesterID, addresseeID uint64) (*domain.Friendship, error) {
	var f domain.Friendship
	err := r.db.WithContext(ctx).
		Where("requester_id = ? AND addressee_id = ?", requesterID, addresseeID).
		First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *friendshipRepository) DeleteBetween(ctx context.Context, requesterID, addresseeID uint64) error {
	return r.db.WithContext(ctx).
		Where("requester_id = ? AND addressee_id = ?", requesterID, addresseeID).
		Delete(&domain.Friendship{}).Error
}
