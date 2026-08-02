package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// profileImagePresigner は profile アイコン用の S3 presigner（profiles/{userId}/{epochNs}{ext} キー）。
type profileImagePresigner struct {
	pre s3Presigner
}

// NewProfileImagePresigner は本番経路。
func NewProfileImagePresigner(p s3Presigner) repository.ProfileImagePresigner {
	return &profileImagePresigner{pre: p}
}

// NewStubProfileImagePresigner は test / dev 用 stub。
func NewStubProfileImagePresigner(bucket string) repository.ProfileImagePresigner {
	return &profileImagePresigner{pre: &stubPresigner{bucket: bucket}}
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
		// 配信ドメインを含めない（FRESTYLE-234）。画像はアプリと同一オリジンで配信され、
		// ブラウザが現在のドメインを補って解決するため、この形ならドメインを変えても
		// 保存済みデータを書き換えずに済む。絶対 URL で保存していた頃は、ドメイン移行の
		// たびに過去の画像が全て参照不能になっていた（FRESTYLE-232 で実害）。
		ImageURL:  "/" + key,
		Key:       key,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

// guessExt は fileName または contentType から拡張子を返す。
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
