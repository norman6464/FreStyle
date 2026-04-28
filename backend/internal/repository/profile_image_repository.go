package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileImagePresigner は profile アイコン用 S3 PUT 署名付き URL を発行する。
// AWS SDK 実装は別 PR。現状は Stub で形式だけ満たす。
type ProfileImagePresigner interface {
	Generate(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error)
}

type stubProfileImagePresigner struct {
	bucket string
	cdnURL string
}

// NewStubProfileImagePresigner は CDN URL (例: https://normanblog.com) と
// S3 bucket name を受け取って stub presigner を返す。
func NewStubProfileImagePresigner(bucket, cdnURL string) ProfileImagePresigner {
	return &stubProfileImagePresigner{bucket: bucket, cdnURL: strings.TrimRight(cdnURL, "/")}
}

func (p *stubProfileImagePresigner) Generate(_ context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = "image/png"
	}
	ext := guessExt(fileName, contentType)
	key := fmt.Sprintf("profiles/%d/%d%s", userID, time.Now().UnixNano(), ext)
	return &domain.ProfileImageUploadURL{
		UploadURL: fmt.Sprintf("https://%s.s3.amazonaws.com/%s?X-Amz-Stub=1", p.bucket, key),
		ImageURL:  fmt.Sprintf("%s/%s", p.cdnURL, key),
		Key:       key,
		ExpiresIn: 600,
	}, nil
}

// guessExt は fileName または contentType から「.png」「.jpg」などの拡張子を返す。
// 純粋関数化することで usecase 経由で間接的にテストできる。
func guessExt(fileName, contentType string) string {
	if i := strings.LastIndex(fileName, "."); i != -1 && i < len(fileName)-1 {
		return strings.ToLower(fileName[i:])
	}
	switch contentType {
	case "image/jpeg", "image/jpg":
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
