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

func (s *stubUserRepo) ListByWorkspaceID(_ context.Context, _ string) ([]domain.User, error) {
	return nil, s.err
}

func (s *stubUserRepo) Create(_ context.Context, _ *domain.User) error {
	return s.err
}

func (s *stubUserRepo) UpdateName(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func (s *stubUserRepo) UpdateWorkspaceID(_ context.Context, _ uint64, _ *string) error {
	return s.err
}

func (s *stubUserRepo) UpdateActive(context.Context, uint64, bool) error { return nil }
func (s *stubUserRepo) SoftDelete(context.Context, uint64) error         { return nil }

func (s *stubUserRepo) FindActiveByEmail(context.Context, string) (*domain.User, error) {
	return nil, nil
}

func (s *stubUserRepo) CognitoSubjectByUserID(context.Context, uint64) (string, error) {
	return "", nil
}

// fakeTxManager は repository.TxManager のテスト用 no-op 実装。fn(ctx) をそのまま呼ぶだけで、
// 実 DB もトランザクションも介さない。package usecase（白箱テスト）内の複数ファイルで共有する。
type fakeTxManager struct{}

func (fakeTxManager) DoInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
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
