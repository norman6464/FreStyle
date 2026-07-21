package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetProfileUseCase は指定 user のプロフィールを返す。
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

// UpdateProfileUseCase はプロフィールの任意フィールドを upsert する。
type UpdateProfileUseCase struct {
	profiles repository.ProfileRepository
}

func NewUpdateProfileUseCase(p repository.ProfileRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{profiles: p}
}

type UpdateProfileInput struct {
	UserID        uint64
	Bio           string
	AvatarURL     string
	StatusMessage string
}

func (u *UpdateProfileUseCase) Execute(ctx context.Context, in UpdateProfileInput) (*domain.Profile, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	// Expand フェーズ: 旧 status と新 status_message の両方へ同値を書く(dual-write)。
	// これにより新旧タスクが同時稼働しても両列が有効な値を保つ。
	p := &domain.Profile{
		UserID:        in.UserID,
		Bio:           in.Bio,
		AvatarURL:     in.AvatarURL,
		Status:        in.StatusMessage,
		StatusMessage: in.StatusMessage,
	}
	if err := u.profiles.Upsert(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}
