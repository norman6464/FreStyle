package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

type ReminderSettingRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*domain.ReminderSetting, error)
	Upsert(ctx context.Context, s *domain.ReminderSetting) error
}

type reminderSettingRepository struct{ db *gorm.DB }

func NewReminderSettingRepository(db *gorm.DB) ReminderSettingRepository {
	return &reminderSettingRepository{db: db}
}

func (r *reminderSettingRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.ReminderSetting, error) {
	var s domain.ReminderSetting
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *reminderSettingRepository) Upsert(ctx context.Context, s *domain.ReminderSetting) error {
	return r.db.WithContext(ctx).Save(s).Error
}
