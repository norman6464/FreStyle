package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// profileImagePresigner は [repository.ProfileImagePresigner] を 満たす S3 プレゼンタ。
// profiles/{userId}/{epochNs}{ext} 形式 の キー で PUT URL を 発行 する。
type profileImagePresigner struct {
	pre    s3Presigner
	cdnURL string
}

// NewProfileImagePresigner は infra/s3.Presigner を受けて real な presigner を返す。
// cdnURL は配信用ベース URL (CloudFront / S3 virtual-hosted-style)。
func NewProfileImagePresigner(p s3Presigner, cdnURL string) repository.ProfileImagePresigner {
	return &profileImagePresigner{pre: p, cdnURL: strings.TrimRight(cdnURL, "/")}
}

// NewStubProfileImagePresigner は test / dev 用の stub。
// 本番では NewProfileImagePresigner(infra/s3.NewPresigner(...), cdnURL) を使う。
func NewStubProfileImagePresigner(bucket, cdnURL string) repository.ProfileImagePresigner {
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
