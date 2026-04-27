package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubBookmarkRepo struct {
	rows []domain.ScenarioBookmark
	err  error
}

func (s *stubBookmarkRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.ScenarioBookmark, error) {
	return s.rows, s.err
}
func (s *stubBookmarkRepo) Create(_ context.Context, b *domain.ScenarioBookmark) error {
	if s.err != nil {
		return s.err
	}
	b.ID = 7
	return nil
}
func (s *stubBookmarkRepo) Delete(_ context.Context, _, _ uint64) error { return s.err }

func TestListScenarioBookmarks_RequiresUserID(t *testing.T) {
	uc := NewListScenarioBookmarksUseCase(&stubBookmarkRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddScenarioBookmark_AssignsID(t *testing.T) {
	uc := NewAddScenarioBookmarkUseCase(&stubBookmarkRepo{})
	got, err := uc.Execute(context.Background(), 1, 2)
	if err != nil || got.ID != 7 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestRemoveScenarioBookmark_RequiresArgs(t *testing.T) {
	uc := NewRemoveScenarioBookmarkUseCase(&stubBookmarkRepo{})
	if err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error")
	}
}
