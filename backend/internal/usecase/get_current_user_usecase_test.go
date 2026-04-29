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
func (s *stubUserRepo) Create(_ context.Context, _ *domain.User) error { return s.err }
func (s *stubUserRepo) UpdateDisplayName(_ context.Context, _ uint64, _ string) error {
	return s.err
}
func (s *stubUserRepo) UpdateRole(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func TestGetCurrentUserUseCase_Found(t *testing.T) {
	want := &domain.User{ID: 1, CognitoSub: "abc", Email: "u@example.com"}
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: want})
	got, err := uc.Execute(context.Background(), "abc")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got == nil || got.ID != 1 {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestGetCurrentUserUseCase_NotFound(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{user: nil})
	got, err := uc.Execute(context.Background(), "missing")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
}

func TestGetCurrentUserUseCase_Error(t *testing.T) {
	uc := NewGetCurrentUserUseCase(&stubUserRepo{err: errors.New("db down")})
	if _, err := uc.Execute(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
