package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListSharedSessionsUseCase struct {
	repo repository.SharedSessionRepository
}

func NewListSharedSessionsUseCase(r repository.SharedSessionRepository) *ListSharedSessionsUseCase {
	return &ListSharedSessionsUseCase{repo: r}
}

func (u *ListSharedSessionsUseCase) Execute(ctx context.Context, limit int) ([]domain.SharedSession, error) {
	return u.repo.ListPublic(ctx, limit)
}

type CreateSharedSessionUseCase struct {
	repo repository.SharedSessionRepository
}

func NewCreateSharedSessionUseCase(r repository.SharedSessionRepository) *CreateSharedSessionUseCase {
	return &CreateSharedSessionUseCase{repo: r}
}

type CreateSharedSessionInput struct {
	OwnerID     uint64
	SessionID   uint64
	Title       string
	Description string
	IsPublic    bool
}

func (u *CreateSharedSessionUseCase) Execute(ctx context.Context, in CreateSharedSessionInput) (*domain.SharedSession, error) {
	if in.OwnerID == 0 || in.SessionID == 0 {
		return nil, errors.New("ownerID and sessionID are required")
	}
	if in.Title == "" {
		return nil, errors.New("title is required")
	}
	s := &domain.SharedSession{
		OwnerID: in.OwnerID, SessionID: in.SessionID,
		Title: in.Title, Description: in.Description, IsPublic: in.IsPublic,
	}
	if err := u.repo.Create(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}
