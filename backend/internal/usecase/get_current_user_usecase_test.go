package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubUserRepo struct {
	user *domain.User
	err  error
}

func (s *stubUserRepo) FindByCognitoSub(_ context.Context, _ string) (*domain.User, error) {
	return s.user, s.err
}

func (s *stubUserRepo) FindByID(_ context.Context, _ uint64) (*domain.User, error) {
	return s.user, s.err
}

func (s *stubUserRepo) ListByRole(_ context.Context, _ domain.RoleName) ([]domain.User, error) {
	return nil, s.err
}

func (s *stubUserRepo) ListByCompanyID(_ context.Context, _ uint64) ([]domain.User, error) {
	return nil, s.err
}

func (s *stubUserRepo) UpdateAiChatEnabled(_ context.Context, _ uint64, _ *bool) error {
	return s.err
}

func (s *stubUserRepo) CreateWithOidcIdentity(_ context.Context, _ *domain.User, _, _ string) error {
	return s.err
}

func (s *stubUserRepo) CreateFirstSuperAdminWithOidcIdentity(
	_ context.Context, _ *domain.User, _, _ string,
) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return true, nil
}

func (s *stubUserRepo) UpdateName(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func (s *stubUserRepo) UpdateRole(_ context.Context, _ uint64, _ domain.RoleName) error {
	return s.err
}

func (s *stubUserRepo) UpdatePlatformAdmin(_ context.Context, _ uint64, _ bool) error {
	return s.err
}

func (s *stubUserRepo) UpdateCompanyID(_ context.Context, _ uint64, _ uint64) error {
	return s.err
}

func (s *stubUserRepo) UpdateActive(context.Context, uint64, bool) error { return nil }
func (s *stubUserRepo) SoftDelete(context.Context, uint64) error         { return nil }

func (s *stubUserRepo) EnsureOidcIdentity(context.Context, uint64, string, string) error {
	return nil
}

func (s *stubUserRepo) FindActiveByEmail(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (s *stubUserRepo) CognitoSubjectByUserID(context.Context, uint64) (string, error) {
	return "", nil
}

func Test_現在ユーザー取得_見つかる(t *testing.T) {
	want := &domain.User{ID: 1, Email: "u@example.com"}
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: want})
	got, err := uc.Execute(context.Background(), "abc")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func Test_現在ユーザー取得_見つからない(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: nil})
	got, err := uc.Execute(context.Background(), "missing")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func Test_現在ユーザー取得_エラー(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{err: errors.New("db down")})
	if _, err := uc.Execute(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
