package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubPresigner struct {
	url    *domain.NoteImageUploadURL
	err    error
	called bool
}

func (s *stubPresigner) Generate(_ context.Context, _ uint64, _ string) (*domain.NoteImageUploadURL, error) {
	s.called = true
	return s.url, s.err
}

func newStubbedNoteImageUseCase() (*IssueNoteImageUploadURLUseCase, *stubPresigner) {
	stub := &stubPresigner{
		url: &domain.NoteImageUploadURL{URL: "https://example", Key: "k", ExpiresIn: 60},
	}
	return NewIssueNoteImageUploadURLUseCase(stub), stub
}

func Test_ノート画像アップロードURL発行_ユーザーIDが必須(t *testing.T) {
	uc, stub := newStubbedNoteImageUseCase()
	_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
		UserID:      0,
		ContentType: "image/png",
		SizeBytes:   1024,
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if stub.called {
		t.Fatal("presigner should not be called")
	}
}

func Test_ノート画像アップロードURL発行_許可MIMEはURLを返す(t *testing.T) {
	for contentType := range AllowedNoteImageContentTypes {
		t.Run(contentType, func(t *testing.T) {
			uc, _ := newStubbedNoteImageUseCase()
			got, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      1,
				ContentType: contentType,
				SizeBytes:   1024,
			})
			if err != nil || got.URL == "" {
				t.Fatalf("unexpected: %+v err=%v", got, err)
			}
		})
	}
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
			uc, stub := newStubbedNoteImageUseCase()
			_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      1,
				ContentType: contentType,
				SizeBytes:   1024,
			})
			if !errors.Is(err, ErrNoteImageUnsupportedType) {
				t.Fatalf("want ErrNoteImageUnsupportedType, got %v", err)
			}
			if stub.called {
				t.Fatal("presigner should not be called")
			}
		})
	}
}

func Test_ノート画像アップロードURL発行_サイズ検証(t *testing.T) {
	cases := []struct {
		name    string
		size    int64
		wantErr error
	}{
		{name: "0 は拒否", size: 0, wantErr: ErrNoteImageTooLarge},
		{name: "負数は拒否", size: -1, wantErr: ErrNoteImageTooLarge},
		{name: "上限ちょうどは許可", size: maxNoteImageBytes, wantErr: nil},
		{name: "上限超過は拒否", size: maxNoteImageBytes + 1, wantErr: ErrNoteImageTooLarge},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uc, stub := newStubbedNoteImageUseCase()
			_, err := uc.Execute(context.Background(), IssueNoteImageUploadURLInput{
				UserID:      1,
				ContentType: "image/png",
				SizeBytes:   tc.size,
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			if stub.called != (tc.wantErr == nil) {
				t.Fatalf("presigner called = %v, want %v", stub.called, tc.wantErr == nil)
			}
		})
	}
}
