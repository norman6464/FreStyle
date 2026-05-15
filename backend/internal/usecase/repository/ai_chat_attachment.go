package repository

import "context"

// AiChatAttachmentPresigner は AI チャットの添付ファイル用 PUT presigned URL を発行する。
//
// note_image / profile_image と同様の責務だが、key prefix と content-type の許容範囲が違う。
// 共通化せず別プレゼンタを切るのは、将来 PR-G2 で document (PDF/CSV) 対応をする際に
// 「画像と document で別バケット / 別 lifecycle にしたい」可能性を残すため。
//
// 実装: [github.com/norman6464/FreStyle/backend/internal/adapter/persistence] の
// aiChatAttachmentPresigner (S3 SDK)。
type AiChatAttachmentPresigner interface {
	Generate(ctx context.Context, userID uint64, filename, contentType string) (*AiChatAttachmentUploadURL, error)
}

// AiChatAttachmentUploadURL は presigned URL 発行結果。
//
// Key は S3 オブジェクトキー。クライアントは UploadURL に PUT した後、
// この Key を SSE 送信ペイロードの attachments[].key に詰めて返す。
type AiChatAttachmentUploadURL struct {
	UploadURL string `json:"uploadUrl"`
	Key       string `json:"key"`
	ExpiresIn int    `json:"expiresIn"`
}
