package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_画像アップロード検証_ContentTypeの許可リスト は許可リストの内外を固定する。
//
// 変異確認: AcceptedImageContentTypes のチェックを外すと、この表の
// 「許可リスト外は拒否」側のケースが緑のまま落ちなくなる（ErrUnsupportedImageContentType
// を返さなくなるため）— このテストがそれを捕まえる。
func Test_画像アップロード検証_ContentTypeの許可リスト(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		wantErr     error
	}{
		{"png は許可", "image/png", nil},
		{"jpeg は許可", "image/jpeg", nil},
		{"gif は許可", "image/gif", nil},
		{"webp は許可", "image/webp", nil},
		{"svg は許可リスト外", "image/svg+xml", domain.ErrUnsupportedImageContentType},
		{"pdf は許可リスト外", "application/pdf", domain.ErrUnsupportedImageContentType},
		{"空文字は許可リスト外", "", domain.ErrUnsupportedImageContentType},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := domain.ValidateImageUpload(c.contentType, 1024)
			if c.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}

// Test_画像アップロード検証_サイズの境界値 は 0・上限ちょうど・上限+1・負数の境界を固定する。
func Test_画像アップロード検証_サイズの境界値(t *testing.T) {
	cases := []struct {
		name    string
		size    int64
		wantErr error
	}{
		{"0 は拒否", 0, domain.ErrImageTooLarge},
		{"負数は拒否", -1, domain.ErrImageTooLarge},
		{"上限ちょうどは許可", domain.MaxImageUploadBytes, nil},
		{"上限+1は拒否", domain.MaxImageUploadBytes + 1, domain.ErrImageTooLarge},
		{"上限未満の通常値は許可", 1024, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := domain.ValidateImageUpload("image/png", c.size)
			if c.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}
