package profile

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubProfileRepo struct {
	p   *domain.Profile
	err error
}

func (s *stubProfileRepo) FindByUserID(_ context.Context, _ uint64) (*domain.Profile, error) {
	return s.p, s.err
}

func (s *stubProfileRepo) Upsert(_ context.Context, p *domain.Profile) error {
	if s.err != nil {
		return s.err
	}
	s.p = p
	return nil
}

func Test_プロフィール取得_ユーザーIDが必須(t *testing.T) {
	uc := NewGetProfileUseCase(&stubProfileRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_プロフィール取得_見つからなければnil(t *testing.T) {
	uc := NewGetProfileUseCase(&stubProfileRepo{p: nil})
	got, err := uc.Execute(context.Background(), 1)
	if err != nil || got != nil {
		t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
	}
}

func Test_プロフィール更新_ユーザーIDが必須(t *testing.T) {
	uc := NewUpdateProfileUseCase(&stubProfileRepo{})
	if _, err := uc.Execute(context.Background(), UpdateProfileInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_プロフィール更新_永続化する(t *testing.T) {
	repo := &stubProfileRepo{}
	uc := NewUpdateProfileUseCase(repo)
	got, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: 1, Bio: "hi"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.Bio != "hi" {
		t.Fatalf("expected bio=hi, got %q", got.Bio)
	}
}

// Contract フェーズ: StatusMessage の入力が status_message へ書かれること(status 列は廃止)。
func Test_プロフィール更新_status_messageに書き込む(t *testing.T) {
	repo := &stubProfileRepo{}
	uc := NewUpdateProfileUseCase(repo)
	if _, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: 1, StatusMessage: "元気です"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.p.StatusMessage != "元気です" {
		t.Fatalf("status_message 未書き込み: %q", repo.p.StatusMessage)
	}
}

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

func Test_プロフィール画像アップロードURL発行_ユーザーIDが必須(t *testing.T) {
	uc := NewIssueProfileImageUploadURLUseCase(&stubProfileImagePresigner{})
	if _, err := uc.Execute(context.Background(), 0, "a.png", "image/png"); err == nil {
		t.Fatal("expected error")
	}
}

func Test_プロフィール画像アップロードURL発行_presignerへ引数を渡す(t *testing.T) {
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
