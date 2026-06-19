package handler

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// nopActivityRepo は handler テスト全体で共用する UserDailyActivityRepository stub。
// 実際の DB 書き込みを行わず、常に nil エラーを返す。
type nopActivityRepo struct{}

func (n *nopActivityRepo) Increment(_ context.Context, _ uint64, _ time.Time, _ repository.UserDailyActivityIncrement) error {
	return nil
}

func (n *nopActivityRepo) ListByUser(_ context.Context, _ uint64, _, _ time.Time) ([]domain.UserDailyActivity, error) {
	return nil, nil
}
