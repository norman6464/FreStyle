package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_チケット添付presigner はアップロード_ダウンロード双方が秒単位の有効期限を返すことを
// 固定する（kbImagePresigner と同じ imagePresigner を共有するラッパーなので、TTL の
// time.Duration → int 秒への変換がここだけの責務）。
func Test_チケット添付presigner(t *testing.T) {
	ctx := context.Background()
	p := NewStubTicketAttachmentPresigner("bucket")

	url, expiresIn, err := p.PresignUpload(ctx, "tickets/ws/ticket/1.bin", "application/pdf", 1024)
	require.NoError(t, err)
	assert.Contains(t, url, "tickets/ws/ticket/1.bin")
	assert.Equal(t, int(10*time.Minute/time.Second), expiresIn)

	url, expiresIn, err = p.PresignDownload(ctx, "tickets/ws/ticket/1.bin")
	require.NoError(t, err)
	assert.Contains(t, url, "tickets/ws/ticket/1.bin")
	assert.Equal(t, int(10*time.Minute/time.Second), expiresIn)
}
