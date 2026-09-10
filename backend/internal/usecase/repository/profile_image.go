package repository

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileImagePresigner は profile アイコン用 PUT 署名付き URL を発行する。
//
// fileName は受け取らない。オブジェクトの拡張子は検査済みの contentType からのみ導く
// （利用者が指定するファイル名から作ると、Content-Type と食い違う拡張子を選べてしまう）。
type ProfileImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string, size int64) (*domain.ProfileImageUploadURL, error)
}
