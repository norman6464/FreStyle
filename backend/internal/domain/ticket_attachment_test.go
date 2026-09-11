package domain_test

import (
	"strings"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_添付ファイル名検証 は形の妥当性を固定する。
//
// 変異確認: len(name) > MaxAttachmentFilenameBytes のチェックを外すと、「上限+1は拒否」の
// ケースが緑のまま落ちなくなる。
func Test_添付ファイル名検証(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		wantErr  error
	}{
		{"通常のファイル名は許可", "資料.pdf", nil},
		{"空文字は拒否", "", domain.ErrInvalidAttachmentFilename},
		{"前後に空白があると拒否", " 資料.pdf", domain.ErrInvalidAttachmentFilename},
		{"上限ちょうどは許可", strings.Repeat("a", domain.MaxAttachmentFilenameBytes), nil},
		{"上限+1は拒否", strings.Repeat("a", domain.MaxAttachmentFilenameBytes+1), domain.ErrInvalidAttachmentFilename},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := domain.ValidateAttachmentFilename(c.filename)
			if c.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, c.wantErr)
		})
	}
}
