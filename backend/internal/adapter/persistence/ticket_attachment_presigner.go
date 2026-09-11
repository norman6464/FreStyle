package persistence

import (
	"context"

	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ticketAttachmentPresigner はチケット添付用 presigner
// （tickets/{workspaceId}/{ticketId}/{epochNs}.bin キー。キーの組み立ては呼び出し側の
// usecase が行う。kbImagePresigner と同じ形 — imagePresigner を共有バケットで多重化する）。
type ticketAttachmentPresigner struct {
	pre imagePresigner
}

// NewTicketAttachmentPresigner は本番経路。infra/gcs.Presigner を渡して使う。
func NewTicketAttachmentPresigner(p imagePresigner) repository.TicketAttachmentPresigner {
	return &ticketAttachmentPresigner{pre: p}
}

// NewStubTicketAttachmentPresigner は test / dev 用 stub。
func NewStubTicketAttachmentPresigner(bucket string) repository.TicketAttachmentPresigner {
	return &ticketAttachmentPresigner{pre: &stubPresigner{bucket: bucket}}
}

func (p *ticketAttachmentPresigner) PresignUpload(ctx context.Context, key, contentType string, size int64) (string, int, error) {
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType, size)
	if err != nil {
		return "", 0, err
	}
	return url, int(ttl.Seconds()), nil
}

func (p *ticketAttachmentPresigner) PresignDownload(ctx context.Context, key string) (string, int, error) {
	url, ttl, err := p.pre.PresignGet(ctx, key)
	if err != nil {
		return "", 0, err
	}
	return url, int(ttl.Seconds()), nil
}
