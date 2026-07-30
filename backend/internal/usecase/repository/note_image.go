package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteImagePresigner は S3 への PUT 用 presigned URL を発行する。
// sizeBytes は Content-Length として署名に焼き込まれ、超過アップロードを S3 側で拒否させる。
type NoteImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string, sizeBytes int64) (*domain.NoteImageUploadURL, error)
}
