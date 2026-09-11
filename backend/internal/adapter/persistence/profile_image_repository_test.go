package persistence

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spyImagePresigner は imagePresigner の spy。PresignPut に渡された引数を記録する
// （呼ばれたこと自体が「検証を通り抜けて署名まで進んだ」ことの証拠になる）。
type spyImagePresigner struct {
	called           bool
	gotKey, gotCType string
	gotContentLength int64
}

func (s *spyImagePresigner) PresignPut(_ context.Context, key, contentType string, contentLength int64) (string, time.Duration, error) {
	s.called = true
	s.gotKey = key
	s.gotCType = contentType
	s.gotContentLength = contentLength
	return "https://storage.googleapis.com/bucket/" + key, 10 * time.Minute, nil
}

func (s *spyImagePresigner) PresignGet(context.Context, string) (string, time.Duration, error) {
	return "", 0, errors.New("not used in this test")
}

func Test_プロフィール画像presigner_不正な入力はpresign前に拒否する(t *testing.T) {
	t.Run("許可リスト外のContent_Typeは署名を発行しない", func(t *testing.T) {
		spy := &spyImagePresigner{}
		p := NewProfileImagePresigner(spy)
		_, err := p.Generate(context.Background(), 7, "text/html", 1024)
		require.ErrorIs(t, err, domain.ErrUnsupportedImageContentType)
		assert.False(t, spy.called, "検証で落ちたら PresignPut を呼ばないはず")
	})

	t.Run("サイズ0以下は署名を発行しない", func(t *testing.T) {
		spy := &spyImagePresigner{}
		p := NewProfileImagePresigner(spy)
		_, err := p.Generate(context.Background(), 7, "image/png", 0)
		require.ErrorIs(t, err, domain.ErrImageTooLarge)
		assert.False(t, spy.called)
	})

	t.Run("上限を超えるサイズは署名を発行しない", func(t *testing.T) {
		spy := &spyImagePresigner{}
		p := NewProfileImagePresigner(spy)
		_, err := p.Generate(context.Background(), 7, "image/png", domain.MaxImageUploadBytes+1)
		require.ErrorIs(t, err, domain.ErrImageTooLarge)
		assert.False(t, spy.called)
	})

	t.Run("userIDが0なら検証より前に拒否する", func(t *testing.T) {
		spy := &spyImagePresigner{}
		p := NewProfileImagePresigner(spy)
		_, err := p.Generate(context.Background(), 0, "image/png", 1024)
		require.Error(t, err)
		assert.False(t, spy.called)
	})
}

func Test_プロフィール画像presigner_正当な入力はサイズを検査済みで署名に渡す(t *testing.T) {
	spy := &spyImagePresigner{}
	p := NewProfileImagePresigner(spy)
	out, err := p.Generate(context.Background(), 7, "image/jpeg", 2048)
	require.NoError(t, err)
	require.NotNil(t, out)

	assert.True(t, spy.called)
	assert.Equal(t, int64(2048), spy.gotContentLength, "検査済みの size がそのまま Content-Length として渡ること")
	assert.Equal(t, "image/jpeg", spy.gotCType)
	assert.True(t, strings.HasPrefix(spy.gotKey, "profiles/7/"), "key は profiles/<userId>/ から始まること: %s", spy.gotKey)
}

func Test_プロフィール画像presigner_拡張子はContent_Typeからだけ決まる(t *testing.T) {
	cases := []struct {
		contentType string
		wantExt     string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
	}
	for _, tt := range cases {
		t.Run(tt.contentType, func(t *testing.T) {
			spy := &spyImagePresigner{}
			p := NewProfileImagePresigner(spy)
			_, err := p.Generate(context.Background(), 7, tt.contentType, 1024)
			require.NoError(t, err)
			assert.True(t, strings.HasSuffix(spy.gotKey, tt.wantExt),
				"key %q は %q で終わるはず（Content-Type から導いた拡張子）", spy.gotKey, tt.wantExt)
		})
	}
}
