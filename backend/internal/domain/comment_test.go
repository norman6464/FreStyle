package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test_コメント本文検証 は comments.body に保存してよい形の境界を固定する。
//
// 変異確認: len(items) == 0 のチェックを外すと「空配列は拒否」のケースが緑のまま
// 落ちなくなる（ErrInvalidCommentBody を返さなくなるため）— このテストがそれを捕まえる。
func Test_コメント本文検証(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "1件の要素を持つ配列はOK", raw: `[{"type":"text","text":"hello"}]`, wantErr: false},
		{name: "複数件の要素を持つ配列もOK", raw: `[{"type":"text","text":"a"},{"type":"text","text":"b"}]`, wantErr: false},
		{name: "textを持たないinlineノード（hardBreak等）はOK", raw: `[{"type":"hardBreak"}]`, wantErr: false},
		{name: "空配列は拒否（本文の無いコメント）", raw: `[]`, wantErr: true},
		{name: "配列でないJSON（object）は拒否", raw: `{"type":"text"}`, wantErr: true},
		{name: "配列でないJSON（string）は拒否", raw: `"hello"`, wantErr: true},
		{name: "不正なJSONは拒否", raw: `not json`, wantErr: true},
		{name: "空文字は拒否", raw: ``, wantErr: true},
		// CodeRabbit 指摘: [null] や [{}] は「非空の配列」チェックだけでは通ってしまい、
		// KbCommentItem が text を持たないノードを無視するため保存後に空コメントとして残る。
		{name: "null要素を含む配列は拒否", raw: `[null]`, wantErr: true},
		{name: "空objectの要素を含む配列は拒否（typeが無い）", raw: `[{}]`, wantErr: true},
		{name: "textノードでtextが空文字は拒否", raw: `[{"type":"text","text":""}]`, wantErr: true},
		{name: "textノードでtextが空白のみは拒否", raw: `[{"type":"text","text":"   "}]`, wantErr: true},
		{name: "配列の要素がobjectでない（数値）は拒否", raw: `[1]`, wantErr: true},
		{name: "有効な要素と無効な要素が混在すれば拒否", raw: `[{"type":"text","text":"ok"},{}]`, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := domain.ValidateCommentBody(tc.raw)
			if tc.wantErr {
				assert.ErrorIs(t, err, domain.ErrInvalidCommentBody)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_コメントスレッド_Resolved(t *testing.T) {
	unresolved := domain.CommentThread{}
	assert.False(t, unresolved.Resolved())

	resolvedAt := unresolved.CreatedAt
	resolved := domain.CommentThread{ResolvedAt: &resolvedAt}
	assert.True(t, resolved.Resolved())
}
