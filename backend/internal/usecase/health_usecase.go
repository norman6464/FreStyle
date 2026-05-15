package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
)

// CheckHealthUseCase はバックエンドの稼働状態を判定するユースケース。
// DB 到達性を確認し、UP/DOWN を返す。
type CheckHealthUseCase struct {
	repo legacyrepository.HealthRepository
}

func NewCheckHealthUseCase(repo legacyrepository.HealthRepository) *CheckHealthUseCase {
	return &CheckHealthUseCase{repo: repo}
}

func (u *CheckHealthUseCase) Execute(ctx context.Context) domain.Health {
	if err := u.repo.PingDB(ctx); err != nil {
		return domain.Health{Status: domain.StatusDown, DBStatus: domain.StatusDown}
	}
	return domain.Health{Status: domain.StatusUp, DBStatus: domain.StatusUp}
}
