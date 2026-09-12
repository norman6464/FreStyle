package profile

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
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

func (s *stubProfileRepo) UpdateStatus(_ context.Context, userID uint64, emoji, text string, expiresAt *time.Time) (*domain.Profile, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.p = &domain.Profile{UserID: userID, StatusEmoji: emoji, StatusText: text, StatusExpiresAt: expiresAt}
	return s.p, nil
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

// StatusText の入力が status_text へ書かれること。
func Test_プロフィール更新_status_textに書き込む(t *testing.T) {
	repo := &stubProfileRepo{}
	uc := NewUpdateProfileUseCase(repo)
	if _, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: 1, StatusText: "元気です"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.p.StatusText != "元気です" {
		t.Fatalf("status_text 未書き込み: %q", repo.p.StatusText)
	}
}

func Test_ステータス更新_ユーザーIDが必須(t *testing.T) {
	uc := NewUpdateStatusUseCase(&stubProfileRepo{})
	if _, err := uc.Execute(context.Background(), UpdateStatusInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_ステータス更新_絵文字テキスト失効時刻を渡す(t *testing.T) {
	repo := &stubProfileRepo{}
	uc := NewUpdateStatusUseCase(repo)
	expires := time.Now().Add(time.Hour)
	got, err := uc.Execute(context.Background(), UpdateStatusInput{UserID: 1, Emoji: "🎉", Text: "休暇中", ExpiresAt: &expires})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.StatusEmoji != "🎉" || got.StatusText != "休暇中" || got.StatusExpiresAt == nil || !got.StatusExpiresAt.Equal(expires) {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

type stubIdentityRepo struct {
	identities []domain.UserIdentity
	err        error
}

func (s *stubIdentityRepo) EnsureIdentity(context.Context, uint64, string, string) error {
	return nil
}

func (s *stubIdentityRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.UserIdentity, error) {
	return s.identities, s.err
}

func Test_認証方法一覧_ユーザーIDが必須(t *testing.T) {
	uc := NewListMyIdentitiesUseCase(&stubIdentityRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func Test_認証方法一覧_repositoryの結果をそのまま返す(t *testing.T) {
	want := []domain.UserIdentity{{Provider: "oidc", Subject: "sub-1", CreatedAt: time.Now()}}
	uc := NewListMyIdentitiesUseCase(&stubIdentityRepo{identities: want})
	got, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 1 || got[0].Provider != "oidc" || got[0].Subject != "sub-1" {
		t.Fatalf("unexpected identities: %+v", got)
	}
}

type stubProfileImagePresigner struct {
	called    bool
	gotUserID uint64
	gotCType  string
	gotSize   int64
	err       error
}

func (s *stubProfileImagePresigner) Generate(_ context.Context, userID uint64, contentType string, size int64) (*domain.ProfileImageUploadURL, error) {
	s.called = true
	s.gotUserID = userID
	s.gotCType = contentType
	s.gotSize = size
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
	if _, err := uc.Execute(context.Background(), 0, "image/png", 1024); err == nil {
		t.Fatal("expected error")
	}
}

func Test_プロフィール画像アップロードURL発行_presignerへ引数を渡す(t *testing.T) {
	stub := &stubProfileImagePresigner{}
	uc := NewIssueProfileImageUploadURLUseCase(stub)
	got, err := uc.Execute(context.Background(), 7, "image/jpeg", 2048)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.UploadURL == "" || got.ImageURL == "" {
		t.Errorf("URLs should be set: %+v", got)
	}
	if !stub.called || stub.gotUserID != 7 || stub.gotCType != "image/jpeg" || stub.gotSize != 2048 {
		t.Errorf("presigner not called with expected args: %+v", stub)
	}
}
