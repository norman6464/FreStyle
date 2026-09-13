package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// richTextImagePresigner はリッチテキスト画像用の presigner（rich-text/{userId}/{epochNs}.bin キー）。
type richTextImagePresigner struct {
	pre imagePresigner
}

// NewRichTextImagePresigner は本番経路。infra/gcs.Presigner を渡して使う。
func NewRichTextImagePresigner(p imagePresigner) repository.RichTextImagePresigner {
	return &richTextImagePresigner{pre: p}
}

// NewStubRichTextImagePresigner は test / dev 用 stub。
func NewStubRichTextImagePresigner(bucket string) repository.RichTextImagePresigner {
	return &richTextImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *richTextImagePresigner) Generate(ctx context.Context, userID uint64, contentType string, size int64) (*domain.RichTextImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	// Content-Type とサイズの検証は presign より前に済ませる（検証を飛ばすと、
	// 上限の無い PUT presigned URL をいくらでも発行できてしまうため）。
	if err := domain.ValidateImageUpload(contentType, size); err != nil {
		return nil, err
	}
	key := fmt.Sprintf("rich-text/%d/%d.bin", userID, time.Now().UnixNano())
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType, size)
	if err != nil {
		return nil, err
	}
	return &domain.RichTextImageUploadURL{
		URL:       url,
		Key:       key,
		PublicURL: "/" + key,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}
