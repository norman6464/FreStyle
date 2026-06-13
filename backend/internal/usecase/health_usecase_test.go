package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubHealthRepo struct {
	err error
}

func (s *stubHealthRepo) PingDB(_ context.Context) error { return s.err }

func Test_ヘルスチェック_DB正常(t *testing.T) {
	uc := NewCheckHealthUseCase(&stubHealthRepo{err: nil})
	got := uc.Execute(context.Background())
	if got.Status != domain.StatusUp || got.DBStatus != domain.StatusUp {
		t.Fatalf("expected UP/UP, got %+v", got)
	}
}

func Test_ヘルスチェック_DB異常(t *testing.T) {
	uc := NewCheckHealthUseCase(&stubHealthRepo{err: errors.New("ping failed")})
	got := uc.Execute(context.Background())
	if got.Status != domain.StatusDown || got.DBStatus != domain.StatusDown {
		t.Fatalf("expected DOWN/DOWN, got %+v", got)
	}
}
