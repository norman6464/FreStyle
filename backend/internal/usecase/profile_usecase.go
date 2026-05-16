package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetProfileUseCase は 指定 user の プロフィール を 返す。 章 002 §3.3 で 解説 する
// DIP の 標準 例 (usecase は port だけ 知って GORM を 知ら ない)。
// 依存 port: [repository.ProfileRepository] (FindByUserID)。
type GetProfileUseCase struct {
	profiles repository.ProfileRepository
}

func NewGetProfileUseCase(p repository.ProfileRepository) *GetProfileUseCase {
	return &GetProfileUseCase{profiles: p}
}

func (u *GetProfileUseCase) Execute(ctx context.Context, userID uint64) (*domain.Profile, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.profiles.FindByUserID(ctx, userID)
}

// UpdateProfileUseCase は プロフィール の 任意 フィールド を upsert する。
// 依存 port: [repository.ProfileRepository] (Upsert)。
type UpdateProfileUseCase struct {
	profiles repository.ProfileRepository
}

func NewUpdateProfileUseCase(p repository.ProfileRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{profiles: p}
}

type UpdateProfileInput struct {
	UserID    uint64
	Bio       string
	AvatarURL string
	Status    string
}

func (u *UpdateProfileUseCase) Execute(ctx context.Context, in UpdateProfileInput) (*domain.Profile, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	p := &domain.Profile{
		UserID:    in.UserID,
		Bio:       in.Bio,
		AvatarURL: in.AvatarURL,
		Status:    in.Status,
	}
	if err := u.profiles.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
