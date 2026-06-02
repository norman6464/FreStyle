// Package persistence は usecase 層が定義した port の永続化実装
// （GORM / DynamoDB / S3 presigner 等）を集約する。wiring は router.go で行う。
package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userRepository は [repository.UserRepository] の GORM 実装。
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
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

func (r *userRepository) ListByRole(ctx context.Context, role string) ([]domain.User, error) {
	var rows []domain.User
	err := r.db.WithContext(ctx).
		Where("role = ? AND deleted_at IS NULL", role).
		Find(&rows).Error
	return rows, err
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

func (r *userRepository) UpdateRole(ctx context.Context, userID uint64, role string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("role", role).Error
}

func (r *userRepository) UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("company_id", companyID).Error
}

func (r *userRepository) MarkOnboarded(ctx context.Context, userID uint64) error {
	// IS NULL ガードで二度押しでも初回日時を保持する（冪等）。
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND onboarded_at IS NULL", userID).
		Update("onboarded_at", now).Error
}
