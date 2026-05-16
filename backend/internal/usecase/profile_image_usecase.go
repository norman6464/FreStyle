package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueProfileImageUploadURLUseCase は profile アイコン 用 S3 PUT 署名 付き URL を 発行 する。
// 章 009 で 解説 する Presigner port の 標準 例。
// 依存 port: [repository.ProfileImagePresigner] (S3 presigned URL 生成)。
type IssueProfileImageUploadURLUseCase struct {
	presigner repository.ProfileImagePresigner
}

func NewIssueProfileImageUploadURLUseCase(p repository.ProfileImagePresigner) *IssueProfileImageUploadURLUseCase {
	return &IssueProfileImageUploadURLUseCase{presigner: p}
}

// Execute は current user の profile アイコン用 PUT 署名付き URL を発行する。
// userID は handler 側で middleware 経由 (=current user) で渡す前提。
func (u *IssueProfileImageUploadURLUseCase) Execute(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.presigner.Generate(ctx, userID, fileName, contentType)
}
