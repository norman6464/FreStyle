package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubPracticeRepo struct {
	rows []domain.PracticeScenario
	one  *domain.PracticeScenario
	err  error
}

func (s *stubPracticeRepo) ListActive(_ context.Context) ([]domain.PracticeScenario, error) {
	return s.rows, s.err
}
func (s *stubPracticeRepo) FindByID(_ context.Context, _ uint64) (*domain.PracticeScenario, error) {
	return s.one, s.err
}

func TestListPracticeScenarios(t *testing.T) {
	uc := NewListPracticeScenariosUseCase(&stubPracticeRepo{rows: []domain.PracticeScenario{{ID: 1}}})
	got, err := uc.Execute(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestGetPracticeScenario_RequiresID(t *testing.T) {
	uc := NewGetPracticeScenarioUseCase(&stubPracticeRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetPracticeScenario_Returns(t *testing.T) {
	uc := NewGetPracticeScenarioUseCase(&stubPracticeRepo{one: &domain.PracticeScenario{ID: 5}})
	got, err := uc.Execute(context.Background(), 5)
	if err != nil || got.ID != 5 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
