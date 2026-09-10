package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_添付アップロード検証_ContentTypeの許可リスト は許可リストの内外を固定する。
//
// 変異確認: AcceptedAttachmentContentTypes のチェックを外すと、この表の
// 「許可リスト外は拒否」側のケースが緑のまま落ちなくなる（ErrUnsupportedAttachmentContentType
// を返さなくなるため）— このテストがそれを捕まえる。
func Test_添付アップロード検証_ContentTypeの許可リスト(t *testing.T) {
	cases := []struct {
		name        string
		contentType string
		wantErr     error
	}{
		{"pdf は許可", "application/pdf", nil},
		{"png は許可", "image/png", nil},
		{"zip は許可", "application/zip", nil},
		{"csv は許可", "text/csv", nil},
		{"docx は許可", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", nil},
		{"svg は許可リスト外（保存型XSSの経路になるため）", "image/svg+xml", domain.ErrUnsupportedAttachmentContentType},
		{"html は許可リスト外（同上）", "text/html", domain.ErrUnsupportedAttachmentContentType},
		{"実行可能形式は許可リスト外", "application/x-msdownload", domain.ErrUnsupportedAttachmentContentType},
		{"空文字は許可リスト外", "", domain.ErrUnsupportedAttachmentContentType},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := domain.ValidateAttachmentUpload(c.contentType, 1024)
			if c.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}

// Test_添付アップロード検証_サイズの境界値 は 0・上限ちょうど・上限+1・負数の境界を固定する。
func Test_添付アップロード検証_サイズの境界値(t *testing.T) {
	cases := []struct {
		name    string
		size    int64
		wantErr error
	}{
		{"0 は拒否", 0, domain.ErrAttachmentTooLarge},
		{"負数は拒否", -1, domain.ErrAttachmentTooLarge},
		{"上限ちょうどは許可", domain.MaxAttachmentUploadBytes, nil},
		{"上限+1は拒否", domain.MaxAttachmentUploadBytes + 1, domain.ErrAttachmentTooLarge},
		{"上限未満の通常値は許可", 1024, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := domain.ValidateAttachmentUpload("application/pdf", c.size)
			if c.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}
