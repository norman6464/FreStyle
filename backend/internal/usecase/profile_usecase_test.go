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

func TestGetProfile_RequiresUserID(t *testing.T) {
	uc := NewGetProfileUseCase(&stubProfileRepo{})
	if _, err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetProfile_NotFoundReturnsNil(t *testing.T) {
	uc := NewGetProfileUseCase(&stubProfileRepo{p: nil})
	got, err := uc.Execute(context.Background(), 1)
	if err != nil || got != nil {
		t.Fatalf("expected (nil,nil), got (%v,%v)", got, err)
	}
}

func TestUpdateProfile_RequiresUserID(t *testing.T) {
	uc := NewUpdateProfileUseCase(&stubProfileRepo{})
	if _, err := uc.Execute(context.Background(), UpdateProfileInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateProfile_Persists(t *testing.T) {
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
