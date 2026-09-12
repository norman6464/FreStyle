package repository

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/domain"
)

// RichTextImagePresigner はオブジェクトストレージへの PUT 用 presigned URL を発行する。
// size はバイト数。Content-Type とあわせて domain.ValidateImageUpload で検証してから presign する
// （許可リストとサイズ上限）。
type RichTextImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string, size int64) (*domain.RichTextImageUploadURL, error)
}
