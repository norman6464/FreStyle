package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteImagePresigner は S3 への PUT 用 presigned URL を発行する。
type NoteImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error)
}
