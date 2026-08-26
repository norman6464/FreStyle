package persistence

import (
	"context"
	"database/sql"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// healthRepository は [repository.HealthRepository] の実装。
// 接続プール（*sql.DB）の PingContext で疎通を確かめる。
type healthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) repository.HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingDB(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
