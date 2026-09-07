package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// RichTextImagePresigner は S3 への PUT 用 presigned URL を発行する。
// size はバイト数。Content-Type とあわせて domain.ValidateImageUpload で検証してから presign する
// （FRESTYLE-9: 許可リストとサイズ上限）。
type RichTextImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string, size int64) (*domain.RichTextImageUploadURL, error)
}
