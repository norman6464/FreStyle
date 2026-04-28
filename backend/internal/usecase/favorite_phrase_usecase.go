package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListFavoritePhrasesUseCase struct{ repo repository.FavoritePhraseRepository }

func NewListFavoritePhrasesUseCase(r repository.FavoritePhraseRepository) *ListFavoritePhrasesUseCase {
	return &ListFavoritePhrasesUseCase{repo: r}
}

func (u *ListFavoritePhrasesUseCase) Execute(ctx context.Context, userID uint64) ([]domain.FavoritePhrase, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type AddFavoritePhraseUseCase struct{ repo repository.FavoritePhraseRepository }

func NewAddFavoritePhraseUseCase(r repository.FavoritePhraseRepository) *AddFavoritePhraseUseCase {
	return &AddFavoritePhraseUseCase{repo: r}
}

func (u *AddFavoritePhraseUseCase) Execute(ctx context.Context, userID uint64, phrase, note string) (*domain.FavoritePhrase, error) {
	if userID == 0 || phrase == "" {
		return nil, errors.New("userID and phrase are required")
	}
	p := &domain.FavoritePhrase{UserID: userID, Phrase: phrase, Note: note}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

type DeleteFavoritePhraseUseCase struct{ repo repository.FavoritePhraseRepository }

func NewDeleteFavoritePhraseUseCase(r repository.FavoritePhraseRepository) *DeleteFavoritePhraseUseCase {
	return &DeleteFavoritePhraseUseCase{repo: r}
}

func (u *DeleteFavoritePhraseUseCase) Execute(ctx context.Context, userID, id uint64) error {
	if userID == 0 {
		return errors.New("userID is required")
	}
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.Delete(ctx, userID, id)
}
