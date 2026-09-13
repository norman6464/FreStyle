package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// profileImagePresigner は profile アイコン用の presigner（profiles/{userId}/{epochNs}{ext} キー）。
type profileImagePresigner struct {
	pre imagePresigner
}

// NewProfileImagePresigner は本番経路。
func NewProfileImagePresigner(p imagePresigner) repository.ProfileImagePresigner {
	return &profileImagePresigner{pre: p}
}

// NewStubProfileImagePresigner は test / dev 用 stub。
func NewStubProfileImagePresigner(bucket string) repository.ProfileImagePresigner {
	return &profileImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *profileImagePresigner) Generate(ctx context.Context, userID uint64, contentType string, size int64) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	// Content-Type とサイズの検証は presign より前に済ませる（rich_text_image_repository.go
	// と同じ形。検証を飛ばすと、任意の Content-Type・上限の無い PUT presigned URL を
	// いくらでも発行できてしまう）。
	if err := domain.ValidateImageUpload(contentType, size); err != nil {
		return nil, err
	}
	key := fmt.Sprintf("profiles/%d/%d%s", userID, time.Now().UnixNano(), extForContentType(contentType))
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType, size)
	if err != nil {
		return nil, err
	}
	return &domain.ProfileImageUploadURL{
		UploadURL: url,
		ImageURL:  "/" + key,
		Key:       key,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

// extForContentType は検査済みの Content-Type から拡張子を返す。利用者が指定する
// ファイル名からは導かない（Content-Type と食い違う拡張子を選べてしまうため）。
// domain.ValidateImageUpload を通った後に呼ぶ前提（domain.AcceptedImageContentTypes に
// 無い値が来ることは無い）。
func extForContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
