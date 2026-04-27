package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetDailyGoalUseCase struct{ repo repository.DailyGoalRepository }

func NewGetDailyGoalUseCase(r repository.DailyGoalRepository) *GetDailyGoalUseCase {
	return &GetDailyGoalUseCase{repo: r}
}

func (u *GetDailyGoalUseCase) Execute(ctx context.Context, userID uint64, date time.Time) (*domain.DailyGoal, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.FindByUserAndDate(ctx, userID, date)
}

type UpsertDailyGoalUseCase struct{ repo repository.DailyGoalRepository }

func NewUpsertDailyGoalUseCase(r repository.DailyGoalRepository) *UpsertDailyGoalUseCase {
	return &UpsertDailyGoalUseCase{repo: r}
}

type UpsertDailyGoalInput struct {
	UserID    uint64
	Date      time.Time
	TargetMin int
	ActualMin int
}

func (u *UpsertDailyGoalUseCase) Execute(ctx context.Context, in UpsertDailyGoalInput) (*domain.DailyGoal, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.TargetMin < 0 || in.ActualMin < 0 {
		return nil, errors.New("minutes must be non-negative")
	}
	g := &domain.DailyGoal{
		UserID: in.UserID, Date: in.Date,
		TargetMin: in.TargetMin, ActualMin: in.ActualMin,
		IsAchieved: in.ActualMin >= in.TargetMin && in.TargetMin > 0,
	}
	if err := u.repo.Upsert(ctx, g); err != nil {
		return nil, err
	}
	return g, nil
}
