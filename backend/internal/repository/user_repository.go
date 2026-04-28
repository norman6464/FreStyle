package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// UserRepository は users テーブルへのアクセスを提供する。
type UserRepository interface {
	FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	// UpdateDisplayName は ProfilePage の「ニックネーム」変更で呼ばれる。
	UpdateDisplayName(ctx context.Context, userID uint64, displayName string) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("cognito_sub = ? AND deleted_at IS NULL", sub).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) UpdateDisplayName(ctx context.Context, userID uint64, displayName string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("display_name", displayName).Error
}
