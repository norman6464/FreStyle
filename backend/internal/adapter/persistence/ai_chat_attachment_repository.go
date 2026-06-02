package persistence

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// aiChatAttachmentPresigner は AI チャット添付用の S3 presigner（ai-chat/{userId}/{uuid}.{ext} キー）。
type aiChatAttachmentPresigner struct {
	pre s3Presigner
}

// NewAiChatAttachmentPresigner は本番経路。infra/s3.Presigner を渡して使う。
func NewAiChatAttachmentPresigner(p s3Presigner) repository.AiChatAttachmentPresigner {
	return &aiChatAttachmentPresigner{pre: p}
}

// NewStubAiChatAttachmentPresigner は test / dev 用 stub。
func NewStubAiChatAttachmentPresigner(bucket string) repository.AiChatAttachmentPresigner {
	return &aiChatAttachmentPresigner{pre: &stubPresigner{bucket: bucket}}
}

// Generate は PUT presigned URL を返す。filename は拡張子推定にだけ使い key には埋めない（衝突 / インジェクション回避）。
func (p *aiChatAttachmentPresigner) Generate(ctx context.Context, userID uint64, filename, contentType string) (*repository.AiChatAttachmentUploadURL, error) {
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
	return &repository.AiChatAttachmentUploadURL{
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
