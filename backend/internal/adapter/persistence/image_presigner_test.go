package persistence

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 表示 URL に配信ドメインを含めないことを固定する（FRESTYLE-234）。
//
// 絶対 URL で保存していると、配信ドメインを変えるたびに過去の画像が全て参照不能になる
// （FRESTYLE-232 で実害）。ここが絶対 URL に戻ると同じ障害が再発するため、
// 「/ で始まりドメインを含まない」ことを契約としてテストで固定する。
func Test_画像の表示URLに配信ドメインを含めないこと(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		publicURL  func(t *testing.T) string
		wantPrefix string
	}{
		{
			name: "プロフィール画像",
			publicURL: func(t *testing.T) string {
				t.Helper()
				out, err := NewStubProfileImagePresigner("bucket").
					Generate(ctx, 7, "icon.png", "image/png")
				require.NoError(t, err)
				return out.ImageURL
			},
			wantPrefix: "/profiles/7/",
		},
		{
			name: "ノート画像",
			publicURL: func(t *testing.T) string {
				t.Helper()
				out, err := NewStubNoteImagePresigner("bucket").Generate(ctx, 7, "image/png")
				require.NoError(t, err)
				return out.PublicURL
			},
			wantPrefix: "/notes/7/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.publicURL(t)

			assert.True(t, strings.HasPrefix(got, tt.wantPrefix),
				"%s で始まるルート相対パスであること: %s", tt.wantPrefix, got)
			assert.NotContains(t, got, "://", "配信ドメイン（スキーム）を含めないこと: %s", got)
			assert.False(t, strings.HasPrefix(got, "//"),
				"//host/path 形式（プロトコル相対）にしないこと: %s", got)
		})
	}
}

// アップロード先の presigned URL は S3 への直接 PUT なので絶対 URL のままでよい。
// 表示 URL と混同して相対化してしまわないことを確認する。
func Test_アップロード先URLは絶対URLのままであること(t *testing.T) {
	ctx := context.Background()

	profile, err := NewStubProfileImagePresigner("bucket").
		Generate(ctx, 7, "icon.png", "image/png")
	require.NoError(t, err)
	assert.Contains(t, profile.UploadURL, "://", "アップロード先は絶対 URL であること")

	note, err := NewStubNoteImagePresigner("bucket").Generate(ctx, 7, "image/png")
	require.NoError(t, err)
	assert.Contains(t, note.URL, "://", "アップロード先は絶対 URL であること")
}

func Test_画像のキーは表示URLから先頭のスラッシュを外した値になること(t *testing.T) {
	ctx := context.Background()

	// 将来キー保存方式へ移す場合に、相対パスから機械的にキーを得られることを担保する。
	profile, err := NewStubProfileImagePresigner("bucket").
		Generate(ctx, 7, "icon.png", "image/png")
	require.NoError(t, err)
	assert.Equal(t, profile.Key, strings.TrimPrefix(profile.ImageURL, "/"))

	note, err := NewStubNoteImagePresigner("bucket").Generate(ctx, 7, "image/png")
	require.NoError(t, err)
	assert.Equal(t, note.Key, strings.TrimPrefix(note.PublicURL, "/"))
}
