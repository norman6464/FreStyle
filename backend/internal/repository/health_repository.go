package repository

import (
	"context"

	"gorm.io/gorm"
)

// HealthRepository は DB の到達性を確認する。
type HealthRepository interface {
	PingDB(ctx context.Context) error
}

type healthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingDB(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
