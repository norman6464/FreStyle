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
	Create(ctx context.Context, user *domain.User) error
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

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}
