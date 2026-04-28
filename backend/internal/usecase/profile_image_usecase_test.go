package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubProfileImagePresigner struct {
	called    bool
	gotUserID uint64
	gotFile   string
	gotCType  string
	err       error
}

func (s *stubProfileImagePresigner) Generate(_ context.Context, userID uint64, fileName, contentType string) (*domain.ProfileImageUploadURL, error) {
	s.called = true
	s.gotUserID = userID
	s.gotFile = fileName
	s.gotCType = contentType
	if s.err != nil {
		return nil, s.err
	}
	return &domain.ProfileImageUploadURL{
		UploadURL: "https://stub.example/upload",
		ImageURL:  "https://stub.example/image",
		Key:       "profiles/x.png",
		ExpiresIn: 600,
	}, nil
}

func TestIssueProfileImageUploadURL_RequiresUserID(t *testing.T) {
	uc := NewIssueProfileImageUploadURLUseCase(&stubProfileImagePresigner{})
	if _, err := uc.Execute(context.Background(), 0, "a.png", "image/png"); err == nil {
		t.Fatal("expected error")
	}
}

func TestIssueProfileImageUploadURL_PassesArgsToPresigner(t *testing.T) {
	stub := &stubProfileImagePresigner{}
	uc := NewIssueProfileImageUploadURLUseCase(stub)
	got, err := uc.Execute(context.Background(), 7, "icon.jpg", "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.UploadURL == "" || got.ImageURL == "" {
		t.Errorf("URLs should be set: %+v", got)
	}
	if !stub.called || stub.gotUserID != 7 || stub.gotFile != "icon.jpg" || stub.gotCType != "image/jpeg" {
		t.Errorf("presigner not called with expected args: %+v", stub)
	}
}
