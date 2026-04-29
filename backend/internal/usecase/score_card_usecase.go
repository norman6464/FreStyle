package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListScoreCardsByUserIDUseCase struct {
	repo repository.ScoreCardRepository
}

func NewListScoreCardsByUserIDUseCase(r repository.ScoreCardRepository) *ListScoreCardsByUserIDUseCase {
	return &ListScoreCardsByUserIDUseCase{repo: r}
}

func (u *ListScoreCardsByUserIDUseCase) Execute(ctx context.Context, userID uint64) ([]domain.ScoreCard, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type CreateScoreCardUseCase struct {
	repo repository.ScoreCardRepository
}

func NewCreateScoreCardUseCase(r repository.ScoreCardRepository) *CreateScoreCardUseCase {
	return &CreateScoreCardUseCase{repo: r}
}

func (u *CreateScoreCardUseCase) Execute(ctx context.Context, card *domain.ScoreCard) (*domain.ScoreCard, error) {
	if card == nil || card.UserID == 0 || card.SessionID == 0 {
		return nil, errors.New("userID and sessionID are required")
	}
	if err := u.repo.Create(ctx, card); err != nil {
		return nil, err
	}
	return card, nil
}
