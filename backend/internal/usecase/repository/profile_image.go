package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileImagePresigner は profile アイコン用 S3 PUT 署名付き URL を発行する。
type ProfileImagePresigner interface {
	Generate(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error)
}
