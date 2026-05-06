package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// CompleteOnboardingUseCase はユーザーが Welcome 画面の「はじめる」を押した時に
// users.onboarded_at を NOW() で更新する。
//
// 仕様:
//   - 既に値が入っているユーザーに対しても 200 を返す（冪等）。repo の MarkOnboarded が
//     IS NULL ガード付きなので二度押しでも初回日時は保持される。
//   - userID は handler で current user から渡す前提（API は自分自身しか変更不可）。
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
