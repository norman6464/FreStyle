package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CompleteOnboardingUseCase は Welcome 完了時に users.onboarded_at を更新する。
// MarkOnboarded が IS NULL ガード付きなので二度押しでも初回日時は保持される（冪等）。
type CompleteOnboardingUseCase struct {
	users repository.UserRepository
}

func NewCompleteOnboardingUseCase(users repository.UserRepository) *CompleteOnboardingUseCase {
	return &CompleteOnboardingUseCase{users: users}
}

func (u *CompleteOnboardingUseCase) Execute(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	return u.users.MarkOnboarded(ctx, userID)
}
