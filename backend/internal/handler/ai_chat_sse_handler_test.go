package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// buildAttachmentsFromRequest はクロスユーザの S3 key を弾く必要がある。
// userID=7 のユーザーが userID=42 配下の key を渡してきたら attachment_key_not_allowed を返す。
func TestBuildAttachmentsFromRequest_RejectsCrossUserKey(t *testing.T) {
	_, err := buildAttachmentsFromRequest(7, []sseAttachmentRequest{
		{
			Key:         "ai-chat/42/abc.png",
			Filename:    "abc.png",
			ContentType: "image/png",
			SizeBytes:   100,
		},
	})
	assert.ErrorContains(t, err, "attachment_key_not_allowed")
}

// 自分の userID 配下の key は正しく domain.Attachment に変換されること。
func TestBuildAttachmentsFromRequest_AcceptsOwnKey(t *testing.T) {
	out, err := buildAttachmentsFromRequest(7, []sseAttachmentRequest{
		{
			Key:         "ai-chat/7/abc.png",
			Filename:    "abc.png",
			ContentType: "image/png",
			SizeBytes:   100,
		},
	})
	assert.NoError(t, err)
	assert.Len(t, out, 1)
	assert.Equal(t, "ai-chat/7/abc.png", out[0].Key)
	assert.Equal(t, "image", out[0].Kind)
	assert.Equal(t, "png", out[0].Format)
}

// MIME 未対応 / 容量超過 / ai-chat 以外の prefix はそれぞれ専用エラーで弾く。
func TestBuildAttachmentsFromRequest_RejectsUnsupportedAndOversize(t *testing.T) {
	cases := []struct {
		name string
		req  sseAttachmentRequest
		want string
	}{
		{
			name: "未対応 MIME",
			req:  sseAttachmentRequest{Key: "ai-chat/7/a.txt", ContentType: "text/plain", SizeBytes: 100},
			want: "attachment_unsupported_type",
		},
		{
			name: "サイズ 0",
			req:  sseAttachmentRequest{Key: "ai-chat/7/a.png", ContentType: "image/png", SizeBytes: 0},
			want: "attachment_too_large",
		},
		{
			name: "サイズ超過",
			req:  sseAttachmentRequest{Key: "ai-chat/7/a.png", ContentType: "image/png", SizeBytes: 10 * 1024 * 1024},
			want: "attachment_too_large",
		},
		{
			name: "ai-chat 外 prefix",
			req:  sseAttachmentRequest{Key: "notes/7/a.png", ContentType: "image/png", SizeBytes: 100},
			want: "attachment_key_not_allowed",
		},
		{
			name: "key 空",
			req:  sseAttachmentRequest{Key: "", ContentType: "image/png", SizeBytes: 100},
			want: "attachment_invalid",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildAttachmentsFromRequest(7, []sseAttachmentRequest{tc.req})
			assert.ErrorContains(t, err, tc.want)
		})
	}
}
