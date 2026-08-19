package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// テストで使う固定値。マジックナンバーを避けるため名前を付ける。
const (
	testNoteImageUserID    uint64 = 7
	testNoteImageSizeBytes int64  = 1024
	testPresignExpiresIn          = 60
	testPresignedURL              = "https://example"
	testPresignedKey              = "notes/7/1.bin"
)

// mockNoteImagePresigner は repository.NoteImagePresigner の testify/mock 実装。
type mockNoteImagePresigner struct {
	mock.Mock
}

func (m *mockNoteImagePresigner) Generate(ctx context.Context, userID uint64, contentType string, sizeBytes int64) (*domain.NoteImageUploadURL, error) {
	args := m.Called(ctx, userID, contentType, sizeBytes)
	url, _ := args.Get(0).(*domain.NoteImageUploadURL)
	return url, args.Error(1)
}

// newMockedNoteImageUseCase は「呼ばれたら成功 URL を返す」mock を注入した usecase を返す。
func newMockedNoteImageUseCase() (*IssueNoteImageUploadURLUseCase, *mockNoteImagePresigner) {
	m := &mockNoteImagePresigner{}
	m.On("Generate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(&domain.NoteImageUploadURL{
			URL:       testPresignedURL,
			Key:       testPresignedKey,
			ExpiresIn: testPresignExpiresIn,
		}, nil).Maybe()
	return NewIssueNoteImageUploadURLUseCase(m), m
}

func Test_ノート画像アップロードURL発行_ユーザーIDが必須(t *testing.T) {
	uc, m := newMockedNoteImageUseCase()
	_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
		UserID:      unsetUserID,
		ContentType: "image/png",
		SizeBytes:   testNoteImageSizeBytes,
	})
	require.Error(t, err)
	m.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func Test_ノート画像アップロードURL発行_許可MIMEはURLを返す(t *testing.T) {
	for contentType := range allowedNoteImageContentTypes {
		t.Run(contentType, func(t *testing.T) {
			uc, _ := newMockedNoteImageUseCase()
			got, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      testNoteImageUserID,
				ContentType: contentType,
				SizeBytes:   testNoteImageSizeBytes,
			})
			require.NoError(t, err)
			require.Equal(t, testPresignedURL, got.URL)
		})
	}
}

// 検証済みの sizeBytes が presigner に渡ることを確認する。
// presign の Content-Length 署名に使われるため、ここが欠けると申告値だけの検証に戻ってしまう。
func Test_ノート画像アップロードURL発行_検証済みサイズをpresignerに渡す(t *testing.T) {
	uc, m := newMockedNoteImageUseCase()
	_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
		UserID:      testNoteImageUserID,
		ContentType: "image/png",
		SizeBytes:   testNoteImageSizeBytes,
	})
	require.NoError(t, err)
	m.AssertCalled(t, "Generate", mock.Anything, testNoteImageUserID, "image/png", testNoteImageSizeBytes)
}

func Test_ノート画像アップロードURL発行_許可外MIMEは拒否(t *testing.T) {
	cases := []string{
		"",
		"text/html",
		"application/javascript",
		"image/svg+xml",
		"application/pdf",
		"application/octet-stream",
		"IMAGE/PNG",
		"image/png; charset=utf-8",
	}
	for _, contentType := range cases {
		t.Run(contentType, func(t *testing.T) {
			uc, m := newMockedNoteImageUseCase()
			_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      testNoteImageUserID,
				ContentType: contentType,
				SizeBytes:   testNoteImageSizeBytes,
			})
			require.ErrorIs(t, err, ErrNoteImageUnsupportedType)
			m.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func Test_ノート画像アップロードURL発行_サイズ検証(t *testing.T) {
	cases := []struct {
		name    string
		size    int64
		wantErr error
	}{
		{name: "0 は拒否", size: 0, wantErr: ErrNoteImageInvalidSize},
		{name: "負数は拒否", size: -1, wantErr: ErrNoteImageInvalidSize},
		{name: "上限ちょうどは許可", size: maxNoteImageBytes, wantErr: nil},
		{name: "上限超過は拒否", size: maxNoteImageBytes + 1, wantErr: ErrNoteImageTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc, m := newMockedNoteImageUseCase()
			_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      testNoteImageUserID,
				ContentType: "image/png",
				SizeBytes:   tc.size,
			})
			if tc.wantErr == nil {
				require.NoError(t, err)
				m.AssertCalled(t, "Generate", mock.Anything, testNoteImageUserID, "image/png", tc.size)
				return
			}
			require.ErrorIs(t, err, tc.wantErr)
			m.AssertNotCalled(t, "Generate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}
