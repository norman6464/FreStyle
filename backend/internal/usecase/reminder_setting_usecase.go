package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type GetReminderSettingUseCase struct{ repo repository.ReminderSettingRepository }

func NewGetReminderSettingUseCase(r repository.ReminderSettingRepository) *GetReminderSettingUseCase {
	return &GetReminderSettingUseCase{repo: r}
}

func (u *GetReminderSettingUseCase) Execute(ctx context.Context, userID uint64) (*domain.ReminderSetting, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.FindByUserID(ctx, userID)
}

type UpsertReminderSettingUseCase struct{ repo repository.ReminderSettingRepository }

func NewUpsertReminderSettingUseCase(r repository.ReminderSettingRepository) *UpsertReminderSettingUseCase {
	return &UpsertReminderSettingUseCase{repo: r}
}

type UpsertReminderSettingInput struct {
	UserID       uint64
	HourLocal    int
	MinuteLocal  int
	WeekdaysMask int
	IsActive     bool
}

func (u *UpsertReminderSettingUseCase) Execute(ctx context.Context, in UpsertReminderSettingInput) (*domain.ReminderSetting, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.HourLocal < 0 || in.HourLocal > 23 || in.MinuteLocal < 0 || in.MinuteLocal > 59 {
		return nil, errors.New("invalid time")
	}
	s := &domain.ReminderSetting{
		UserID: in.UserID, HourLocal: in.HourLocal, MinuteLocal: in.MinuteLocal,
		WeekdaysMask: in.WeekdaysMask, IsActive: in.IsActive,
	}
	if err := u.repo.Upsert(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}
