package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubReminderRepo struct {
	s   *domain.ReminderSetting
	err error
}

func (st *stubReminderRepo) FindByUserID(_ context.Context, _ uint64) (*domain.ReminderSetting, error) {
	return st.s, st.err
}
func (st *stubReminderRepo) Upsert(_ context.Context, s *domain.ReminderSetting) error {
	if st.err != nil {
		return st.err
	}
	st.s = s
	return nil
}

func TestGetReminderSetting_RequiresUserID(t *testing.T) {
	uc := NewGetReminderSettingUseCase(&stubReminderRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpsertReminderSetting_TimeValidation(t *testing.T) {
	uc := NewUpsertReminderSettingUseCase(&stubReminderRepo{})
	if _, err := uc.Execute(context.Background(), UpsertReminderSettingInput{UserID: 1, HourLocal: 24}); err == nil {
		t.Fatal("expected error for hour=24")
	}
	if _, err := uc.Execute(context.Background(), UpsertReminderSettingInput{UserID: 1, MinuteLocal: 60}); err == nil {
		t.Fatal("expected error for minute=60")
	}
}

func TestUpsertReminderSetting_OK(t *testing.T) {
	uc := NewUpsertReminderSettingUseCase(&stubReminderRepo{})
	got, err := uc.Execute(context.Background(), UpsertReminderSettingInput{
		UserID: 1, HourLocal: 21, MinuteLocal: 30, WeekdaysMask: 0x7F, IsActive: true,
	})
	if err != nil || got.HourLocal != 21 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
