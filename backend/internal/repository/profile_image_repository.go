package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ProfileImagePresigner は profile アイコン用 S3 PUT 署名付き URL を発行する。
type ProfileImagePresigner interface {
	Generate(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error)
}

// s3Presigner は infra/s3.Presigner と同等の interface を minimal に切り出した型。
// repository が infra/s3 パッケージに直接依存しないよう dep direction を反転 (依存性逆転原則)。
type s3Presigner interface {
	PresignPut(ctx context.Context, key, contentType string) (url string, ttl time.Duration, err error)
}

type profileImagePresigner struct {
	pre    s3Presigner
	cdnURL string
}

// NewProfileImagePresigner は infra/s3.Presigner を受けて real な presigner を返す。
// cdnURL は配信用ベース URL (CloudFront / S3 virtual-hosted-style)。
func NewProfileImagePresigner(p s3Presigner, cdnURL string) ProfileImagePresigner {
	return &profileImagePresigner{pre: p, cdnURL: strings.TrimRight(cdnURL, "/")}
}

// NewStubProfileImagePresigner は test / dev 用の stub。
// 本番では NewProfileImagePresigner(infra/s3.NewPresigner(...), cdnURL) を使う。
func NewStubProfileImagePresigner(bucket, cdnURL string) ProfileImagePresigner {
	return &profileImagePresigner{
		pre:    &stubPresigner{bucket: bucket},
		cdnURL: strings.TrimRight(cdnURL, "/"),
	}
}

func (p *profileImagePresigner) Generate(ctx context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = "image/png"
	}
	ext := guessExt(fileName, contentType)
	key := fmt.Sprintf("profiles/%d/%d%s", userID, time.Now().UnixNano(), ext)
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType)
	if err != nil {
		return nil, err
	}
	return &domain.ProfileImageUploadURL{
		UploadURL: url,
		ImageURL:  fmt.Sprintf("%s/%s", p.cdnURL, key),
		Key:       key,
		ExpiresIn: int(ttl.Seconds()),
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
