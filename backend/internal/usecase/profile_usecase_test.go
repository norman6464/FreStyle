package usecase

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

// Expand フェーズの dual-write を検証: StatusMessage の入力が status と status_message の
// 両列(Profile の両フィールド)へ同値で書かれること。
func Test_プロフィール更新_Expand期はstatusとstatus_messageに二重書きする(t *testing.T) {
	repo := &stubProfileRepo{}
	uc := NewUpdateProfileUseCase(repo)
	if _, err := uc.Execute(context.Background(), UpdateProfileInput{UserID: 1, StatusMessage: "元気です"}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.p.Status != "元気です" || repo.p.StatusMessage != "元気です" {
		t.Fatalf("dual-write 失敗: status=%q status_message=%q", repo.p.Status, repo.p.StatusMessage)
	}
}
