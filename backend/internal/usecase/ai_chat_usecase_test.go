package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// nopActivityRepo は UserDailyActivityRepository の何もしない stub。
type nopActivityRepo struct{}

func (n *nopActivityRepo) Increment(_ context.Context, _ uint64, _ time.Time, _ repository.UserDailyActivityIncrement) error {
	return nil
}

func (n *nopActivityRepo) ListByUser(_ context.Context, _ uint64, _, _ time.Time) ([]domain.UserDailyActivity, error) {
	return nil, nil
}

type stubAiChatSessionRepo struct {
	rows []domain.AiChatSession
	err  error
}

func (s *stubAiChatSessionRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.AiChatSession, error) {
	return s.rows, s.err
}

func (s *stubAiChatSessionRepo) FindByID(_ context.Context, _ uint64) (*domain.AiChatSession, error) {
	return nil, nil
}

func (s *stubAiChatSessionRepo) Create(_ context.Context, sess *domain.AiChatSession) error {
	if s.err != nil {
		return s.err
	}
	sess.ID = 42
	return nil
}

func (s *stubAiChatSessionRepo) UpdateTitle(_ context.Context, _ uint64, _ string) error {
	return s.err
}

func (s *stubAiChatSessionRepo) Delete(_ context.Context, _ uint64) error {
	return s.err
}

func Test_AIチャットセッション一覧_ユーザー別(t *testing.T) {
	repo := &stubAiChatSessionRepo{rows: []domain.AiChatSession{{ID: 1, UserID: 7, Title: "a"}}}
	uc := NewGetAiChatSessionsByUserIDUseCase(repo)
	got, err := uc.Execute(context.Background(), 7)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 session, got %d", len(got))
	}
}

func Test_AIチャットセッション作成_タイトルが必須(t *testing.T) {
	uc := NewCreateAiChatSessionUseCase(&stubAiChatSessionRepo{}, &nopActivityRepo{})
	_, err := uc.Execute(context.Background(), CreateAiChatSessionInput{UserID: 1})
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func Test_AIチャットセッション作成_種別は既定でfree(t *testing.T) {
	uc := NewCreateAiChatSessionUseCase(&stubAiChatSessionRepo{}, &nopActivityRepo{})
	got, err := uc.Execute(context.Background(), CreateAiChatSessionInput{UserID: 1, Title: "x"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got.SessionType != domain.AiChatSessionTypeFree {
		t.Fatalf("want type=free, got %s", got.SessionType)
	}
	if got.ID != 42 {
		t.Fatalf("expected stub-assigned id 42, got %d", got.ID)
	}
}
