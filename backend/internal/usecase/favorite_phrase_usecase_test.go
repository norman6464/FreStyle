package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubFavPhraseRepo struct {
	rows          []domain.FavoritePhrase
	err           error
	deletedUserID uint64
	deletedID     uint64
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
func (s *stubFavPhraseRepo) Delete(_ context.Context, userID, id uint64) error {
	s.deletedUserID = userID
	s.deletedID = id
	return s.err
}

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

func TestDeleteFavoritePhrase_RequiresUserID(t *testing.T) {
	uc := NewDeleteFavoritePhraseUseCase(&stubFavPhraseRepo{})
	if err := uc.Execute(context.Background(), 0, 1); err == nil {
		t.Fatal("expected error when userID is 0")
	}
}

func TestDeleteFavoritePhrase_RequiresID(t *testing.T) {
	uc := NewDeleteFavoritePhraseUseCase(&stubFavPhraseRepo{})
	if err := uc.Execute(context.Background(), 1, 0); err == nil {
		t.Fatal("expected error when id is 0")
	}
}

func TestDeleteFavoritePhrase_PassesUserIDToRepo(t *testing.T) {
	repo := &stubFavPhraseRepo{}
	uc := NewDeleteFavoritePhraseUseCase(repo)
	if err := uc.Execute(context.Background(), 9, 11); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if repo.deletedUserID != 9 || repo.deletedID != 11 {
		t.Fatalf("Delete should be called with (9,11), got (%d,%d)", repo.deletedUserID, repo.deletedID)
	}
}
