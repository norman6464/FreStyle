package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubFavPhraseRepo struct {
	rows []domain.FavoritePhrase
	err  error
}

func (s *stubFavPhraseRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.FavoritePhrase, error) {
	return s.rows, s.err
}
func (s *stubFavPhraseRepo) Create(_ context.Context, p *domain.FavoritePhrase) error {
	if s.err != nil {
		return s.err
	}
	p.ID = 71
	return nil
}
func (s *stubFavPhraseRepo) Delete(_ context.Context, _ uint64) error { return s.err }

func TestListFavoritePhrases_RequiresUserID(t *testing.T) {
	uc := NewListFavoritePhrasesUseCase(&stubFavPhraseRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddFavoritePhrase_RequiresPhrase(t *testing.T) {
	uc := NewAddFavoritePhraseUseCase(&stubFavPhraseRepo{})
	if _, err := uc.Execute(context.Background(), 1, "", ""); err == nil {
		t.Fatal("expected error")
	}
}

func TestAddFavoritePhrase_AssignsID(t *testing.T) {
	uc := NewAddFavoritePhraseUseCase(&stubFavPhraseRepo{})
	got, err := uc.Execute(context.Background(), 1, "hello", "")
	if err != nil || got.ID != 71 {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}

func TestDeleteFavoritePhrase_RequiresID(t *testing.T) {
	uc := NewDeleteFavoritePhraseUseCase(&stubFavPhraseRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}
