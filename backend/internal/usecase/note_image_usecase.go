package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// IssueNoteImageUploadURLUseCase はノート画像用 S3 PUT 署名付き URL を発行する。
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

// maxNoteImageBytes は 1 枚あたりの上限。AI チャット添付の画像上限（maxImageBytes）と同値に揃えている。
// 同一プロダクト内で画像 1 枚の上限が複数あると利用側が混乱するため、数値の根拠より一致を優先した。
const maxNoteImageBytes int64 = 5 * 1024 * 1024

// allowedNoteImageContentTypes は presigned URL 発行を許可する画像 MIME 一覧。
// contentType は presign に焼き込まれ PUT 時のヘッダと一致が要求されるため、
// ここを通った MIME 以外のオブジェクトは共有バケットに置けない（HTML / JS の配信を防ぐ）。
// 許可リストはセキュリティ境界なので、外部パッケージから書き換えられないよう unexport にしている。
var allowedNoteImageContentTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/jpg":  {},
	"image/gif":  {},
	"image/webp": {},
}

// ErrNoteImageUnsupportedType は未対応 MIME。
var ErrNoteImageUnsupportedType = errors.New("note image: unsupported content type")

// ErrNoteImageTooLarge は sizeBytes が上限を超えたことを表す。
var ErrNoteImageTooLarge = errors.New("note image: sizeBytes exceeds the 5MB limit")

func (u *IssueNoteImageUploadURLUseCase) Execute(ctx context.Context, in IssueNoteImageUploadURLInput) (*domain.NoteImageUploadURL, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if _, ok := allowedNoteImageContentTypes[in.ContentType]; !ok {
		return nil, ErrNoteImageUnsupportedType
	}
	if in.SizeBytes <= 0 || in.SizeBytes > maxNoteImageBytes {
		return nil, ErrNoteImageTooLarge
	}
	return u.presigner.Generate(ctx, in.UserID, in.ContentType, in.SizeBytes)
}
