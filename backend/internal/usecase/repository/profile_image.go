package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileImagePresigner は profile アイコン用 S3 PUT 署名付き URL を発行する。
//
// 単一 メソッド の port な ので Effective Go 流 の -er 命名 を 採用。 実装 は
// [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// profileImagePresigner (S3 SDK)。
type ProfileImagePresigner interface {
	Generate(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error)
}
