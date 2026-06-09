// Package persistence は usecase 層が定義した port の永続化実装
// （GORM / DynamoDB / S3 presigner 等）を集約する。wiring は router.go で行う。
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// userRepository は [repository.UserRepository] の実装。
// 読み取りは sqlc 生成コード（生 SQL 直書き）、書き込みは GORM（autoTime / 採番）を使う。
// 接続は GORM の *sql.DB を sqlc に共有する（別 pool を持たない）。
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

// toDomainUser は sqlc 生成モデル → domain への詰め替え。
// id 系は DB が bigint(int64) で domain が uint64。値は採番シーケンス由来で常に非負・int64 範囲内のため
// 変換は安全（gosec G115 は persistence の id 境界として .golangci.yml で除外）。
func toDomainUser(row sqlcgen.User) *domain.User {
	u := &domain.User{
		ID:          uint64(row.ID),
		CognitoSub:  row.CognitoSub,
		Email:       row.Email,
		DisplayName: row.DisplayName,
		Role:        row.Role,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if row.CompanyID.Valid {
		cid := uint64(row.CompanyID.Int64)
		u.CompanyID = &cid
	}
	if row.OnboardedAt.Valid {
		t := row.OnboardedAt.Time
		u.OnboardedAt = &t
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		u.DeletedAt = &t
	}
	return u
}

func (r *userRepository) FindByCognitoSub(ctx context.Context, sub string) (*domain.User, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetUserByCognitoSub(ctx, sub)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	id64, ok := toInt64ID(id)
	if !ok {
		return nil, nil // int64 範囲外 = 存在し得ない id
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	row, err := sqlcgen.New(sqlDB).GetUserByID(ctx, id64)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *userRepository) ListByRole(ctx context.Context, role string) ([]domain.User, error) {
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListUsersByRole(ctx, role)
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *toDomainUser(row))
	}
	return users, nil
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
