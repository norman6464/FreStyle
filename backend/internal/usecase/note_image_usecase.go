package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type IssueNoteImageUploadURLUseCase struct {
	presigner repository.NoteImagePresigner
}

func NewIssueNoteImageUploadURLUseCase(p repository.NoteImagePresigner) *IssueNoteImageUploadURLUseCase {
	return &IssueNoteImageUploadURLUseCase{presigner: p}
}

func (u *IssueNoteImageUploadURLUseCase) Execute(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.presigner.Generate(ctx, userID, contentType)
}
