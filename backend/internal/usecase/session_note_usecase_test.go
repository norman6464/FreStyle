package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

type stubSessionNoteRepo struct {
	n   *domain.SessionNote
	err error
	// upsertCalls は Upsert が実際に呼ばれた回数。
	// 「所有者でない書き込みは repository まで届かない」ことを確かめるために数える
	// （書き込みが起きなかったことは、返り値のエラーだけでは区別が付かないため）。
	upsertCalls int
}

func (s *stubSessionNoteRepo) FindBySessionID(_ context.Context, _ uint64) (*domain.SessionNote, error) {
	return s.n, s.err
}

func (s *stubSessionNoteRepo) Upsert(_ context.Context, n *domain.SessionNote) error {
	s.upsertCalls++
	if s.err != nil {
		return s.err
	}
	s.n = n
	return nil
}

// stubSessionOwnerRepo は所有者検証だけを目的にした AiChatSessionRepository の stub。
// FindByID は本物（persistence 実装）と同じく、未存在を domain.ErrNotFound で返す契約にする。
type stubSessionOwnerRepo struct {
	session *domain.AiChatSession
	err     error
}

func (s *stubSessionOwnerRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.AiChatSession, error) {
	return nil, nil
}

func (s *stubSessionOwnerRepo) FindByID(_ context.Context, _ uint64) (*domain.AiChatSession, error) {
	if s.err != nil {
		return nil, s.err
	}
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

// メモの所有権はセッションの所有権に従属する。書き込んでよいかを決めるのは
// ai_chat_sessions.user_id であって、session_notes.user_id ではない。
//
// SQL の ON CONFLICT ... WHERE は衝突（既存行の UPDATE）のときしか効かないので、
// メモがまだ無いセッションへの初回 INSERT はそこでは止まらない。ここで止める。
func Test_セッションノート保存_他人のセッションには書けない(t *testing.T) {
	t.Run("メモがまだ無くても新規作成できない", func(t *testing.T) {
		notes := &stubSessionNoteRepo{} // メモは未作成（FindBySessionID は nil を返す）
		uc := NewUpsertSessionNoteUseCase(notes, ownedSession(1, 2))

		got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 999, Content: "攻撃者が作成"})
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want domain.ErrNotFound, got %v", err)
		}
		if got != nil {
			t.Fatalf("メモを返してはいけない: %+v", got)
		}
		if notes.upsertCalls != 0 {
			t.Fatalf("repository まで書き込みが届いている: upsertCalls=%d", notes.upsertCalls)
		}
		if notes.n != nil {
			t.Fatalf("行が作られている: %+v", notes.n)
		}
	})

	t.Run("既にメモがある場合も更新できない", func(t *testing.T) {
		notes := &stubSessionNoteRepo{n: &domain.SessionNote{SessionID: 1, UserID: 2, Content: "所有者のメモ"}}
		uc := NewUpsertSessionNoteUseCase(notes, ownedSession(1, 2))

		if _, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 999, Content: "攻撃者が上書き"}); !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("want domain.ErrNotFound, got %v", err)
		}
		if notes.upsertCalls != 0 {
			t.Fatalf("repository まで書き込みが届いている: upsertCalls=%d", notes.upsertCalls)
		}
		if notes.n.Content != "所有者のメモ" {
			t.Fatalf("内容が変わっている: %q", notes.n.Content)
		}
	})
}

// 存在しないセッションと他人のセッションは同じ応答にする。
// 分けると応答差から session_id の実在を総当たりで判別できてしまう。
func Test_セッションノート保存_存在しないセッションはnot_found(t *testing.T) {
	notes := &stubSessionNoteRepo{}
	uc := NewUpsertSessionNoteUseCase(notes, &stubSessionOwnerRepo{}) // FindByID は domain.ErrNotFound

	if _, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 12345, UserID: 2, Content: "x"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want domain.ErrNotFound, got %v", err)
	}
	if notes.upsertCalls != 0 {
		t.Fatalf("repository まで書き込みが届いている: upsertCalls=%d", notes.upsertCalls)
	}
}

// 自分のセッションなら、メモの有無にかかわらず書ける（塞ぎすぎていないことの確認）。
func Test_セッションノート保存_自分のセッションには書ける(t *testing.T) {
	t.Run("メモが無い場合は新規作成できる", func(t *testing.T) {
		notes := &stubSessionNoteRepo{}
		uc := NewUpsertSessionNoteUseCase(notes, ownedSession(1, 2))

		got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 2, Content: "初回"})
		if err != nil || got == nil || got.Content != "初回" {
			t.Fatalf("unexpected: %+v err=%v", got, err)
		}
		if notes.upsertCalls != 1 {
			t.Fatalf("want upsertCalls=1, got %d", notes.upsertCalls)
		}
	})

	t.Run("メモがある場合は更新できる", func(t *testing.T) {
		notes := &stubSessionNoteRepo{n: &domain.SessionNote{SessionID: 1, UserID: 2, Content: "旧"}}
		uc := NewUpsertSessionNoteUseCase(notes, ownedSession(1, 2))

		got, err := uc.Execute(context.Background(), UpsertSessionNoteInput{SessionID: 1, UserID: 2, Content: "新"})
		if err != nil || got == nil || got.Content != "新" {
			t.Fatalf("unexpected: %+v err=%v", got, err)
		}
		if notes.upsertCalls != 1 {
			t.Fatalf("want upsertCalls=1, got %d", notes.upsertCalls)
		}
	})
}
