package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueRichTextImageUploadURLUseCase はリッチテキスト画像用 S3 PUT 署名付き URL を発行する。
type IssueRichTextImageUploadURLUseCase struct {
	presigner repository.RichTextImagePresigner
}

func NewIssueRichTextImageUploadURLUseCase(p repository.RichTextImagePresigner) *IssueRichTextImageUploadURLUseCase {
	return &IssueRichTextImageUploadURLUseCase{presigner: p}
}

func (u *IssueRichTextImageUploadURLUseCase) Execute(ctx context.Context, userID uint64, contentType string) (*domain.RichTextImageUploadURL, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.presigner.Generate(ctx, userID, contentType)
}
