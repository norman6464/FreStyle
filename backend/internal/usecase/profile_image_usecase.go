package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueProfileImageUploadURLUseCase は profile アイコン用 S3 PUT 署名付き URL を発行する。
type IssueProfileImageUploadURLUseCase struct {
	presigner repository.ProfileImagePresigner
}

func NewIssueProfileImageUploadURLUseCase(p repository.ProfileImagePresigner) *IssueProfileImageUploadURLUseCase {
	return &IssueProfileImageUploadURLUseCase{presigner: p}
}

func (u *IssueProfileImageUploadURLUseCase) Execute(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.presigner.Generate(ctx, userID, fileName, contentType)
}
