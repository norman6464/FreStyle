package usecase

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubSessionNoteRepo struct {
	n   *domain.SessionNote
	err error
}

func (s *stubSessionNoteRepo) FindBySessionID(_ context.Context, _ uint64) (*domain.SessionNote, error) {
	return s.n, s.err
}

func (s *stubSessionNoteRepo) Upsert(_ context.Context, n *domain.SessionNote) error {
	if s.err != nil {
		return s.err
	}
	s.n = n
	return nil
}

func Test_セッションノート取得_セッションIDとユーザーIDが必須(t *testing.T) {
	uc := NewGetSessionNoteUseCase(&stubSessionNoteRepo{})

	if _, err := uc.Execute(context.Background(), GetSessionNoteInput{UserID: 2}); err == nil {
		t.Fatal("sessionID 未指定はエラーであるべき")
	}
	if _, err := uc.Execute(context.Background(), GetSessionNoteInput{SessionID: 1}); err == nil {
		t.Fatal("userID 未指定はエラーであるべき")
	}
}

// sessionID を総当たりしても他人のノートは読めない（IDOR の回帰防止）。
func Test_セッションノート取得_他人のノートは返さない(t *testing.T) {
	repo := &stubSessionNoteRepo{n: &domain.SessionNote{SessionID: 1, UserID: 2, Content: "secret"}}
	uc := NewGetSessionNoteUseCase(repo)

	t.Run("所有者本人は取得できる", func(t *testing.T) {
		got, err := uc.Execute(context.Background(), GetSessionNoteInput{SessionID: 1, UserID: 2})
		if err != nil || got == nil || got.Content != "secret" {
			t.Fatalf("unexpected: %+v err=%v", got, err)
		}
	})

	t.Run("別ユーザーには存在しない扱い", func(t *testing.T) {
		got, err := uc.Execute(context.Background(), GetSessionNoteInput{SessionID: 1, UserID: 999})
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if got != nil {
			t.Fatalf("他人のノートが漏れている: %+v", got)
		}
	})
}

// stubSessionOwnerRepo は所有者検証だけを目的にした AiChatSessionRepository の stub。
// FindByID は本物（persistence 実装）と同じく、未存在を domain.ErrNotFound で返す契約にする。
type stubSessionOwnerRepo struct {
	session *domain.AiChatSession
}

func (s *stubSessionOwnerRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.AiChatSession, error) {
	return nil, nil
}

func (s *stubSessionOwnerRepo) FindByID(_ context.Context, _ uint64) (*domain.AiChatSession, error) {
	if s.session == nil {
		return nil, domain.ErrNotFound
	}
	return s.session, nil
}

func (s *stubSessionOwnerRepo) Create(_ context.Context, _ *domain.AiChatSession) error { return nil }

func (s *stubSessionOwnerRepo) UpdateTitle(_ context.Context, _ uint64, _ string) error { return nil }

func (s *stubSessionOwnerRepo) Delete(_ context.Context, _ uint64) error { return nil }

// ownedSession は userID が所有する sessionID のセッションを返すヘルパ。
func ownedSession(sessionID, userID uint64) *stubSessionOwnerRepo {
	return &stubSessionOwnerRepo{session: &domain.AiChatSession{ID: sessionID, UserID: userID}}
}

func Test_セッションノート保存_バリデーション(t *testing.T) {
	uc := NewUpsertSessionNoteUseCase(&stubSessionNoteRepo{}, ownedSession(1, 2))
	if _, err := uc.Execute(context.Background(), UpsertSessionNoteInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_セッションノート保存_永続化する(t *testing.T) {
	repo := &stubSessionNoteRepo{}
	uc := NewUpsertSessionNoteUseCase(repo, ownedSession(1, 2))
	got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 2, Content: "hello"})
	if err != nil || got.Content != "hello" {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
