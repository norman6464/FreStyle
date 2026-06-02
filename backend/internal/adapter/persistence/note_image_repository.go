package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// noteImagePresigner は note 画像用の S3 presigner（notes/{userId}/{epochNs}.bin キー）。
type noteImagePresigner struct {
	pre s3Presigner
}

// NewNoteImagePresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewNoteImagePresigner(p s3Presigner) repository.NoteImagePresigner {
	return &noteImagePresigner{pre: p}
}

// NewStubNoteImagePresigner は test / dev 用 stub。
func NewStubNoteImagePresigner(bucket string) repository.NoteImagePresigner {
	return &noteImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *noteImagePresigner) Generate(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = "image/png"
	}
	key := fmt.Sprintf("notes/%d/%d.bin", userID, time.Now().UnixNano())
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType)
	if err != nil {
		return nil, err
	}
	return &domain.NoteImageUploadURL{
		URL:       url,
		Key:       key,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}
