package usecase

import (
	"context"
	"errors"
	"testing"
)

type stubMarkOnboardedRepo struct {
	stubUserRepo
	calledID uint64
	err      error
}

func (s *stubMarkOnboardedRepo) UpdateActive(context.Context, uint64, bool) error { return nil }
func (s *stubMarkOnboardedRepo) SoftDelete(context.Context, uint64) error         { return nil }
func (s *stubMarkOnboardedRepo) MarkOnboarded(_ context.Context, userID uint64) error {
	s.calledID = userID
	return s.err
}

func Test_オンボーディング完了_ユーザーIDが必須(t *testing.T) {
	uc := NewCompleteOnboardingUseCase(&stubMarkOnboardedRepo{})
	if err := uc.Execute(context.Background(), 0); err == nil {
		t.Fatal("expected error for userID=0")
	}
}

func Test_オンボーディング完了_リポジトリへ委譲(t *testing.T) {
	repo := &stubMarkOnboardedRepo{}
	uc := NewCompleteOnboardingUseCase(repo)
	if err := uc.Execute(context.Background(), 42); err != nil {
		t.Fatalf("err: %v", err)
	}
	if repo.calledID != 42 {
		t.Errorf("expected MarkOnboarded(42), got %d", repo.calledID)
	}
}

func Test_オンボーディング完了_リポジトリエラーを伝播(t *testing.T) {
	uc := NewCompleteOnboardingUseCase(&stubMarkOnboardedRepo{err: errors.New("db down")})
	if err := uc.Execute(context.Background(), 1); err == nil {
		t.Fatal("expected error to bubble")
	}
}
