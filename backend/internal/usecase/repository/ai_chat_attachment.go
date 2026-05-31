package repository

import "context"

// AiChatAttachmentPresigner は AI チャット添付用の PUT presigned URL を発行する。
type AiChatAttachmentPresigner interface {
	Generate(ctx context.Context, userID uint64, filename, contentType string) (*AiChatAttachmentUploadURL, error)
}

// AiChatAttachmentUploadURL は presigned URL 発行結果。
// クライアントは UploadURL に PUT 後、Key を SSE 送信ペイロードの attachments[].key に詰める。
type AiChatAttachmentUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
