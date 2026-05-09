package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// UserRepository は users テーブルへのアクセスを提供する。
type UserRepository interface {
	FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	// UpdateDisplayName は ProfilePage の「氏名」変更、 および OIDC ログイン時に
	// 旧 displayName=email を id_token の name claim で自動補正するときに呼ばれる。
	UpdateDisplayName(ctx context.Context, userID uint64, displayName string) error
	// UpdateRole は Cognito group → DB role 同期、または招待受諾時に呼ばれる。
	UpdateRole(ctx context.Context, userID uint64, role string) error
	// UpdateCompanyID は既存ユーザーが招待を受けて company に紐付くときに呼ばれる。
	UpdateCompanyID(ctx context.Context, userID uint64, companyID uint64) error
	// MarkOnboarded は Welcome 画面の「はじめる」ボタン押下時に呼ばれ、
	// onboarded_at = NOW() に更新する。冪等（既に値が入っていても上書きしない）。
	MarkOnboarded(ctx context.Context, userID uint64) error
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
	// 冪等性: 既に onboarded_at が入っているレコードには上書きしない（IS NULL でガード）。
	// これで「Welcome 画面に戻ってもう一度押された」場合でも初回日時を保持できる。
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ? AND onboarded_at IS NULL", userID).
		Update("onboarded_at", now).Error
}
