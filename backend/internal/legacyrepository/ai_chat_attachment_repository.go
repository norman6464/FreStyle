package legacyrepository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AiChatAttachmentPresigner は AI チャットの添付ファイル用 PUT presigned URL を発行する。
//
// note_image / profile_image と同様の責務だが、key prefix と content-type の許容範囲が違う。
// 共通化せず別プレゼンタを切るのは、将来 PR-G2 で document (PDF/CSV) 対応をする際に
// 「画像と document で別バケット / 別 lifecycle にしたい」可能性を残すため。
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

type aiChatAttachmentPresigner struct {
	pre s3Presigner
}

// NewAiChatAttachmentPresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewAiChatAttachmentPresigner(p s3Presigner) AiChatAttachmentPresigner {
	return &aiChatAttachmentPresigner{pre: p}
}

// NewStubAiChatAttachmentPresigner は test / dev 用 stub。
func NewStubAiChatAttachmentPresigner(bucket string) AiChatAttachmentPresigner {
	return &aiChatAttachmentPresigner{pre: &stubPresigner{bucket: bucket}}
}

// Generate は ai-chat/{userId}/{uuid}.{ext} のキーで PUT presigned URL を返す。
//
// filename は拡張子推定 / オブジェクト名表示用にだけ使い、key には埋めない（衝突回避と
// インジェクション対策）。
func (p *aiChatAttachmentPresigner) Generate(ctx context.Context, userID uint64, filename, contentType string) (*AiChatAttachmentUploadURL, error) {
	if userID == 0 {
		return nil, fmt.Errorf("userID is required")
	}
	if contentType == "" {
		return nil, fmt.Errorf("contentType is required")
	}
	ext := extensionForContentType(contentType, filename)
	key := fmt.Sprintf("ai-chat/%d/%s%s", userID, uuid.New().String(), ext)
	url, ttl, err := p.pre.PresignPut(ctx, key, contentType)
	if err != nil {
		return nil, err
	}
	return &AiChatAttachmentUploadURL{
		UploadURL: url,
		Key:       key,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

// extensionForContentType は MIME から拡張子を逆引きする。
// 不明な MIME は filename 末尾の拡張子をそのまま使う。両方無ければ "" を返す。
func extensionForContentType(contentType, filename string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	case "text/csv":
		return ".csv"
	}
	if i := strings.LastIndexByte(filename, '.'); i >= 0 && i < len(filename)-1 {
		return strings.ToLower(filename[i:])
	}
	return ""
}
