package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteImagePresigner は S3 への PUT 用 presigned URL を発行する。
type NoteImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error)
}

// noteImagePresigner は infra/s3.Presigner を介して real な PUT presign を発行する。
type noteImagePresigner struct {
	pre s3Presigner
}

// NewNoteImagePresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewNoteImagePresigner(p s3Presigner) NoteImagePresigner {
	return &noteImagePresigner{pre: p}
}

// NewStubNoteImagePresigner は test / dev 用 stub。
// 本番では NewNoteImagePresigner(infra/s3.NewPresigner(...)) を使うこと。
func NewStubNoteImagePresigner(bucket string) NoteImagePresigner {
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

// stubPresigner は profile_image / note_image 双方の test / dev 用に共有される。
// 本番経路では infra/s3.NewPresigner が s3Presigner を実装する。
type stubPresigner struct{ bucket string }

func (s *stubPresigner) PresignPut(_ context.Context, key, _ string) (string, time.Duration, error) {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1", s.bucket, key), 10 * time.Minute, nil
}
