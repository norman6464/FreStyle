package usecase

import (
	"context"
	"errors"
	"fmt"

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

// unsetUserID は middleware が current user を解決できなかったことを表す番兵値。
const unsetUserID uint64 = 0

// allowedNoteImageContentTypes は presigned URL 発行を許可する画像 MIME 一覧。
// contentType は presign に焼き込まれ PUT 時のヘッダと一致が要求されるため、
// ここを通った MIME 以外のオブジェクトは共有バケットに置けない（HTML / JS の配信を防ぐ）。
// 許可リストはセキュリティ境界なので、外部パッケージから書き換えられないよう unexport にしている。
//
// AI チャット添付・プロフィール画像と同じ 5 種に揃えており、以下は意図的に除外している。
//   - image/svg+xml: SVG は XML で script や on* 属性を埋め込めるため、CDN から配信されると
//     保存型 XSS の媒体になる。画像形式の中で唯一スクリプトを実行しうる
//   - image/heic, image/heif: iPhone の標準形式だが Safari 以外は表示できず、許可しても
//     「アップロードできるが表示されない」状態になる。対応するなら JPEG 変換が別途必要
//   - image/bmp, image/tiff: 前者は非圧縮で上限に達しやすく、後者はブラウザが表示できない
//
// image/jpg は非標準（ブラウザは image/jpeg を送る）だが、AI チャット添付の許可リストが
// 受け入れているため互換目的で残している。
var allowedNoteImageContentTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/jpg":  {},
	"image/gif":  {},
	"image/webp": {},
}

// ErrNoteImageUnsupportedType は未対応 MIME（handler は 415 を返す）。
var ErrNoteImageUnsupportedType = errors.New("note image: unsupported content type")

// ErrNoteImageInvalidSize は sizeBytes が正数でないことを表す（handler は 400 を返す）。
var ErrNoteImageInvalidSize = errors.New("note image: sizeBytes must be positive")

// ErrNoteImageTooLarge は sizeBytes が上限を超えたことを表す（handler は 413 を返す）。
var ErrNoteImageTooLarge = fmt.Errorf("note image: sizeBytes exceeds the limit of %d bytes", maxNoteImageBytes)

func (u *IssueNoteImageUploadURLUseCase) Execute(ctx context.Context, in IssueNoteImageUploadURLInput) (*domain.NoteImageUploadURL, error) {
	if in.UserID == unsetUserID {
		return nil, errors.New("userID is required")
	}
	if _, ok := allowedNoteImageContentTypes[in.ContentType]; !ok {
		return nil, ErrNoteImageUnsupportedType
	}
	if in.SizeBytes <= 0 {
		return nil, ErrNoteImageInvalidSize
	}
	if in.SizeBytes > maxNoteImageBytes {
		return nil, ErrNoteImageTooLarge
	}
	return u.presigner.Generate(ctx, in.UserID, in.ContentType, in.SizeBytes)
}
