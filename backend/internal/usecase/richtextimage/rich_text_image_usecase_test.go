package richtextimage

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubPresigner struct {
	url *domain.RichTextImageUploadURL
	err error
	// gotSize は Generate に実際に渡された size を記録する（Execute の転送を検証するため。
	// 破棄すると size を渡し忘れても気づけないテストになる — CodeRabbit 指摘）。
	gotSize int64
}

func (s *stubPresigner) Generate(_ context.Context, _ uint64, _ string, size int64) (*domain.RichTextImageUploadURL, error) {
	s.gotSize = size
	return s.url, s.err
}

func Test_リッチテキスト画像アップロードURL発行_ユーザーIDが必須(t *testing.T) {
	uc := NewIssueRichTextImageUploadURLUseCase(&stubPresigner{})
	if _, err := uc.Execute(context.Background(), 0, "image/png", 1024); err == nil {
		t.Fatal("expected error")
	}
}

func Test_リッチテキスト画像アップロードURL発行_URLを返す(t *testing.T) {
	presigner := &stubPresigner{
		url: &domain.RichTextImageUploadURL{URL: "https://example", Key: "k", ExpiresIn: 60},
	}
	uc := NewIssueRichTextImageUploadURLUseCase(presigner)
	got, err := uc.Execute(context.Background(), 1, "image/png", 1024)
	if err != nil || got.URL == "" {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
	if presigner.gotSize != 1024 {
		t.Fatalf("size が Generate に転送されていない: got=%d want=1024", presigner.gotSize)
	}
}
