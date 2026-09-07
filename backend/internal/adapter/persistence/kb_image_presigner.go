package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// kbImagePresigner はナレッジのページ画像用 S3 presigner
// （kb/{workspaceId}/{pageId}/{epochNs}.bin キー。キーの組み立ては呼び出し側の usecase が行う）。
type kbImagePresigner struct {
	pre s3Presigner
}

// NewKbImagePresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewKbImagePresigner(p s3Presigner) repository.KbImagePresigner {
	return &kbImagePresigner{pre: p}
}

// NewStubKbImagePresigner は test / dev 用 stub。
func NewStubKbImagePresigner(bucket string) repository.KbImagePresigner {
	return &kbImagePresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *kbImagePresigner) PresignUpload(ctx context.Context, key, contentType string, size int64) (string, int, error) {
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType, size)
	if err != nil {
		return "", 0, err
	}
	return url, int(ttl.Seconds()), nil
}

func (p *kbImagePresigner) PresignDownload(ctx context.Context, key string) (string, int, error) {
	url, ttl, err := p.pre.PresignGet(ctx, key)
	if err != nil {
		return "", 0, err
	}
	return url, int(ttl.Seconds()), nil
}
