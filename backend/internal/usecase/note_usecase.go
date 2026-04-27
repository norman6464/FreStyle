package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListNotesByUserIDUseCase struct{ repo repository.NoteRepository }

func NewListNotesByUserIDUseCase(r repository.NoteRepository) *ListNotesByUserIDUseCase {
	return &ListNotesByUserIDUseCase{repo: r}
}

func (u *ListNotesByUserIDUseCase) Execute(ctx context.Context, userID uint64) ([]domain.Note, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type CreateNoteUseCase struct{ repo repository.NoteRepository }

func NewCreateNoteUseCase(r repository.NoteRepository) *CreateNoteUseCase {
	return &CreateNoteUseCase{repo: r}
}

type CreateNoteInput struct {
	UserID   uint64
	Title    string
	Content  string
	IsPublic bool
}

func (u *CreateNoteUseCase) Execute(ctx context.Context, in CreateNoteInput) (*domain.Note, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if in.Title == "" {
		return nil, errors.New("title is required")
	}
	n := &domain.Note{UserID: in.UserID, Title: in.Title, Content: in.Content, IsPublic: in.IsPublic}
	if err := u.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

type UpdateNoteUseCase struct{ repo repository.NoteRepository }

func NewUpdateNoteUseCase(r repository.NoteRepository) *UpdateNoteUseCase {
	return &UpdateNoteUseCase{repo: r}
}

type UpdateNoteInput struct {
	ID       uint64
	Title    string
	Content  string
	IsPublic bool
}

func (u *UpdateNoteUseCase) Execute(ctx context.Context, in UpdateNoteInput) (*domain.Note, error) {
	if in.ID == 0 {
		return nil, errors.New("id is required")
	}
	existing, err := u.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	existing.Title = in.Title
	existing.Content = in.Content
	existing.IsPublic = in.IsPublic
	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

type DeleteNoteUseCase struct{ repo repository.NoteRepository }

func NewDeleteNoteUseCase(r repository.NoteRepository) *DeleteNoteUseCase {
	return &DeleteNoteUseCase{repo: r}
}

func (u *DeleteNoteUseCase) Execute(ctx context.Context, id uint64) error {
	if id == 0 {
		return errors.New("id is required")
	}
	return u.repo.Delete(ctx, id)
}
