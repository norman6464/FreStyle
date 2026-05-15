package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteImagePresigner は S3 への PUT 用 presigned URL を発行する。
//
// 単一 メソッド の port な ので Effective Go 流 の -er 命名 を 採用。 実装 は
// [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// noteImagePresigner (S3 SDK)。
type NoteImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error)
}
