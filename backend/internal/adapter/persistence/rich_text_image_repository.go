package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// richTextImagePresigner はリッチテキスト画像用の S3 presigner（rich-text/{userId}/{epochNs}.bin キー）。
type richTextImagePresigner struct {
	pre s3Presigner
}

// NewRichTextImagePresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewRichTextImagePresigner(p s3Presigner) repository.RichTextImagePresigner {
	return &richTextImagePresigner{pre: p}
}

// NewStubRichTextImagePresigner は test / dev 用 stub。
func NewStubRichTextImagePresigner(bucket string) repository.RichTextImagePresigner {
	return &richTextImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *richTextImagePresigner) Generate(ctx context.Context, userID uint64, contentType string) (*domain.RichTextImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = "image/png"
	}
	key := fmt.Sprintf("rich-text/%d/%d.bin", userID, time.Now().UnixNano())
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType)
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
