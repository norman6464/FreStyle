package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetCurrentWeeklyChallengeUseCase struct{ repo repository.WeeklyChallengeRepository }

func NewGetCurrentWeeklyChallengeUseCase(r repository.WeeklyChallengeRepository) *GetCurrentWeeklyChallengeUseCase {
	return &GetCurrentWeeklyChallengeUseCase{repo: r}
}

func (u *GetCurrentWeeklyChallengeUseCase) Execute(ctx context.Context) (*domain.WeeklyChallenge, error) {
	return u.repo.CurrentChallenge(ctx)
}

type CompleteWeeklyChallengeUseCase struct{ repo repository.WeeklyChallengeRepository }

func NewCompleteWeeklyChallengeUseCase(r repository.WeeklyChallengeRepository) *CompleteWeeklyChallengeUseCase {
	return &CompleteWeeklyChallengeUseCase{repo: r}
}

func (u *CompleteWeeklyChallengeUseCase) Execute(ctx context.Context, userID, challengeID uint64) error {
	if userID == 0 || challengeID == 0 {
		return errors.New("userID and challengeID are required")
	}
	p := &domain.WeeklyChallengeProgress{UserID: userID, ChallengeID: challengeID, Completed: true}
	return u.repo.UpsertProgress(ctx, p)
}
