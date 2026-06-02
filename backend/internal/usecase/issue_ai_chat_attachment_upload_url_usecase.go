package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueAiChatAttachmentUploadURLUseCase は AI チャット添付の S3 PUT presigned URL を発行する。
// contentType / sizeBytes を usecase 層で検証し、不正な添付は presigned URL 発行前に弾く。
type IssueAiChatAttachmentUploadURLUseCase struct {
	presigner repository.AiChatAttachmentPresigner
}

func NewIssueAiChatAttachmentUploadURLUseCase(p repository.AiChatAttachmentPresigner) *IssueAiChatAttachmentUploadURLUseCase {
	return &IssueAiChatAttachmentUploadURLUseCase{presigner: p}
}

// IssueAiChatAttachmentUploadURLInput は handler から受け取るリクエスト形。
type IssueAiChatAttachmentUploadURLInput struct {
	UserID      uint64
	Filename    string
	ContentType string
	SizeBytes   int64
}

const (
	maxImageBytes    int64 = 5 * 1024 * 1024
	maxDocumentBytes int64 = 4_500_000 // Bedrock document upper bound
)

// AllowedAttachmentContentTypes は presigned URL 発行 / SSE 送信が許可される MIME 一覧。
var AllowedAttachmentContentTypes = map[string]struct {
	Kind   string // "image" | "document"
	Format string // Bedrock format (png/jpeg/gif/webp/pdf/csv)
	Max    int64
}{
	"image/png":  {Kind: "image", Format: "png", Max: maxImageBytes},
	"image/jpeg": {Kind: "image", Format: "jpeg", Max: maxImageBytes},
	"image/jpg":  {Kind: "image", Format: "jpeg", Max: maxImageBytes},
	"image/gif":  {Kind: "image", Format: "gif", Max: maxImageBytes},
	"image/webp": {Kind: "image", Format: "webp", Max: maxImageBytes},
}

// ErrAttachmentUnsupportedType は未対応 MIME。
var ErrAttachmentUnsupportedType = errors.New("attachment: unsupported content type")

// ErrAttachmentTooLarge はサイズ上限超過。
var ErrAttachmentTooLarge = errors.New("attachment: file too large")

func (u *IssueAiChatAttachmentUploadURLUseCase) Execute(ctx context.Context, in IssueAiChatAttachmentUploadURLInput) (*repository.AiChatAttachmentUploadURL, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	rule, ok := AllowedAttachmentContentTypes[in.ContentType]
	if !ok {
		return nil, ErrAttachmentUnsupportedType
	}
	if in.SizeBytes <= 0 || in.SizeBytes > rule.Max {
		return nil, ErrAttachmentTooLarge
	}
	return u.presigner.Generate(ctx, in.UserID, in.Filename, in.ContentType)
}

// maxDocumentBytes は document 対応で使う予約定数（未使用警告抑止）。
var _ = maxDocumentBytes
