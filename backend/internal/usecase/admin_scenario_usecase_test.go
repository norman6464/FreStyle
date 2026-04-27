package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubAdminScenarioRepo struct {
	rows []domain.PracticeScenario
	err  error
}

func (s *stubAdminScenarioRepo) List(_ context.Context) ([]domain.PracticeScenario, error) {
	return s.rows, s.err
}
func (s *stubAdminScenarioRepo) Create(_ context.Context, sc *domain.PracticeScenario) error {
	if s.err != nil {
		return s.err
	}
	sc.ID = 101
	return nil
}
func (s *stubAdminScenarioRepo) Update(_ context.Context, _ *domain.PracticeScenario) error {
	return s.err
}
func (s *stubAdminScenarioRepo) Delete(_ context.Context, _ uint64) error { return s.err }

func TestListAdminScenarios(t *testing.T) {
	uc := NewListAdminScenariosUseCase(&stubAdminScenarioRepo{rows: []domain.PracticeScenario{{ID: 1}}})
	got, err := uc.Execute(context.Background())
	if err != nil || len(got) != 1 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestCreateAdminScenario_RequiresTitle(t *testing.T) {
	uc := NewCreateAdminScenarioUseCase(&stubAdminScenarioRepo{})
	if _, err := uc.Execute(context.Background(), &domain.PracticeScenario{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateAdminScenario_AssignsID(t *testing.T) {
	uc := NewCreateAdminScenarioUseCase(&stubAdminScenarioRepo{})
	got, err := uc.Execute(context.Background(), &domain.PracticeScenario{Title: "x"})
	if err != nil || got.ID != 101 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestUpdateAdminScenario_RequiresID(t *testing.T) {
	uc := NewUpdateAdminScenarioUseCase(&stubAdminScenarioRepo{})
	if _, err := uc.Execute(context.Background(), &domain.PracticeScenario{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteAdminScenario_RequiresID(t *testing.T) {
	uc := NewDeleteAdminScenarioUseCase(&stubAdminScenarioRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
