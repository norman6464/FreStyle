package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// NoteImagePresigner は S3 への PUT 用 presigned URL を発行する。
// AWS SDK 連携は Phase 11.1 で実装予定。Phase 11 ではモック生成のみ。
type NoteImagePresigner interface {
	Generate(ctx context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error)
}

type stubPresigner struct{ bucket string }

func NewStubNoteImagePresigner(bucket string) NoteImagePresigner {
	return &stubPresigner{bucket: bucket}
}

func (p *stubPresigner) Generate(_ context.Context, userID uint64, contentType string) (*domain.NoteImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = "image/png"
	}
	key := fmt.Sprintf("notes/%d/%d.bin", userID, time.Now().UnixNano())
	return &domain.NoteImageUploadURL{
		URL:       fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1", p.bucket, key),
		Key:       key,
		ExpiresIn: 600,
	}, nil
}
