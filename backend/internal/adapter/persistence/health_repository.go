package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// healthRepository は [repository.HealthRepository] の GORM 実装。
type healthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) repository.HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingDB(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
