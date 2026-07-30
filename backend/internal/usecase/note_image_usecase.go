package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueNoteImageUploadURLUseCase はノート画像用 S3 PUT 署名付き URL を発行する。
// contentType / sizeBytes を usecase 層で検証し、画像以外・上限超過は presigned URL 発行前に弾く。
type IssueNoteImageUploadURLUseCase struct {
	presigner repository.NoteImagePresigner
}

func NewIssueNoteImageUploadURLUseCase(p repository.NoteImagePresigner) *IssueNoteImageUploadURLUseCase {
	return &IssueNoteImageUploadURLUseCase{presigner: p}
}

// IssueNoteImageUploadURLInput は handler から受け取るリクエスト形。
type IssueNoteImageUploadURLInput struct {
	UserID      uint64
	ContentType string
	SizeBytes   int64
}

// maxNoteImageBytes は 1 枚あたりの上限（AI チャット添付の画像上限に合わせる）。
const maxNoteImageBytes int64 = 5 * 1024 * 1024

// AllowedNoteImageContentTypes は presigned URL 発行を許可する画像 MIME 一覧。
// contentType は presign に焼き込まれ PUT 時のヘッダと一致が要求されるため、
// ここを通った MIME 以外のオブジェクトは共有バケットに置けない（HTML / JS の配信を防ぐ）。
var AllowedNoteImageContentTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/jpg":  {},
	"image/gif":  {},
	"image/webp": {},
}

// ErrNoteImageUnsupportedType は未対応 MIME。
var ErrNoteImageUnsupportedType = errors.New("note image: unsupported content type")

// ErrNoteImageTooLarge はサイズ上限超過。
var ErrNoteImageTooLarge = errors.New("note image: file too large")

func (u *IssueNoteImageUploadURLUseCase) Execute(ctx context.Context, in IssueNoteImageUploadURLInput) (*domain.NoteImageUploadURL, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if _, ok := AllowedNoteImageContentTypes[in.ContentType]; !ok {
		return nil, ErrNoteImageUnsupportedType
	}
	if in.SizeBytes <= 0 || in.SizeBytes > maxNoteImageBytes {
		return nil, ErrNoteImageTooLarge
	}
	return u.presigner.Generate(ctx, in.UserID, in.ContentType)
}
