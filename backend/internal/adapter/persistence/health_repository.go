package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// healthRepository は [repository.HealthRepository] の実装。
// GORM からは接続プール（*sql.DB）だけを借り、その PingContext で疎通を確かめる。
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
