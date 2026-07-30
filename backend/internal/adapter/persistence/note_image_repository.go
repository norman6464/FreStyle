package persistence

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// defaultNoteImageContentType は contentType 未指定時のフォールバック。
// usecase が許可リストで検証済みのため通常は到達しないが、他の呼び出し元に備えた既定値。
const defaultNoteImageContentType = "image/png"

// noteImageKeyFormat は S3 オブジェクトキーの組み立て（userID, epoch ナノ秒の順）。
// ユーザーごとに prefix を分け、同一ユーザー内は時刻で一意にする。
const noteImageKeyFormat = "notes/%d/%d.bin"

// noteImagePresigner は note 画像用の S3 presigner（notes/{userId}/{epochNs}.bin キー）。
// cdnBase は配信用 CDN URL（NOTE_IMAGES_CDN_URL）。アップロード後の参照 URL 組み立てに使う。
type noteImagePresigner struct {
	pre     s3Presigner
	cdnBase string
}

// NewNoteImagePresigner は本番経路。infra/s3.Presigner と配信 CDN base を渡して使う。
func NewNoteImagePresigner(p s3Presigner, cdnBase string) repository.NoteImagePresigner {
	return &noteImagePresigner{pre: p, cdnBase: cdnBase}
}

// NewStubNoteImagePresigner は test / dev 用 stub（CDN 未設定で相対 URL を返す）。
func NewStubNoteImagePresigner(bucket string) repository.NoteImagePresigner {
	return &noteImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *noteImagePresigner) Generate(ctx context.Context, userID uint64, contentType string, sizeBytes int64) (*domain.NoteImageUploadURL, error) {
	if userID == unsetUserID {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		contentType = defaultNoteImageContentType
	}
	key := fmt.Sprintf(noteImageKeyFormat, userID, time.Now().UnixNano())
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType, sizeBytes)
	if err != nil {
		return nil, err
	}
	return &domain.NoteImageUploadURL{
		URL:       url,
		Key:       key,
		PublicURL: p.publicURL(key),
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

// publicURL は配信用 URL を組み立てる。cdnBase 未設定時は同一オリジンの相対 URL にフォールバック。
func (p *noteImagePresigner) publicURL(key string) string {
	if p.cdnBase == "" {
		return "/" + key
	}
	return strings.TrimRight(p.cdnBase, "/") + "/" + key
}
