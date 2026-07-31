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

func Test_セッションノート保存_バリデーション(t *testing.T) {
	uc := NewUpsertSessionNoteUseCase(&stubSessionNoteRepo{})
	if _, err := uc.Execute(context.Background(), UpsertSessionNoteInput{}); err == nil {
		t.Fatal("expected error")
	}
}

func Test_セッションノート保存_永続化する(t *testing.T) {
	repo := &stubSessionNoteRepo{}
	uc := NewUpsertSessionNoteUseCase(repo)
	got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 2, Content: "hello"})
	if err != nil || got.Content != "hello" {
		t.Fatalf("unexpected: %+v err=%v", got, err)
	}
}
