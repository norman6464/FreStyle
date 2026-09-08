package domain

import "errors"

// AcceptedImageContentTypes は画像アップロードで受け付ける Content-Type の許可リスト。
// ここに無い型は ValidateImageUpload が ErrUnsupportedImageContentType で弾く
// （FRESTYLE-9: 以前は contentType を検証せずそのまま S3 の PUT presign に焼き込んでいた）。
var AcceptedImageContentTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/gif":  true,
	"image/webp": true,
}

// MaxImageUploadBytes は画像アップロード 1 件あたりのサイズ上限（10 MiB）。
const MaxImageUploadBytes = 10 * 1024 * 1024

// ErrUnsupportedImageContentType は許可リストに無い Content-Type が指定されたときに返す。
var ErrUnsupportedImageContentType = errors.New("unsupported image content type")

// ErrImageTooLarge は size が 0 以下、または上限を超えるときに返す。
var ErrImageTooLarge = errors.New("image is too large")

// ValidateImageUpload は画像アップロード要求（Content-Type とサイズ）が保存してよい形かを見る。
// presign の前に必ず通す入口で、不正なら presign 自体を行わない
// （不正な値のまま署名を発行すると、上限の無い PUT を許すのと同じ穴になる）。
func ValidateImageUpload(contentType string, size int64) error {
	if !AcceptedImageContentTypes[contentType] {
		return ErrUnsupportedImageContentType
	}
	if size <= 0 || size > MaxImageUploadBytes {
		return ErrImageTooLarge
	}
	return nil
}
