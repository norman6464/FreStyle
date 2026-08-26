package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ai_chat_handler のガード分岐（401 / 400）を zero-value handler で検証する。
// いずれも usecase 到達前に早期 return するため nil usecase で安全。

func Test_AIチャットハンドラ_セッション一覧_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "")
	(&AiChatHandler{}).GetSessions(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション作成_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&AiChatHandler{}).CreateSession(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション作成_不正なJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&AiChatHandler{}).CreateSession(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション取得_不正なID(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "abc")
	(&AiChatHandler{}).GetSession(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッションタイトル更新_不正なID(t *testing.T) {
	w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 0, "abc")
	(&AiChatHandler{}).UpdateSessionTitle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッションタイトル更新_タイトル欠落(t *testing.T) {
	w, c := noteCtx(http.MethodPut, `{}`, 0, "5")
	(&AiChatHandler{}).UpdateSessionTitle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション削除_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodDelete, "", 0, "5")
	(&AiChatHandler{}).DeleteSession(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

// fakeAiChatSessionRepo は repository.AiChatSessionRepository の最小 fake。
// 「1 行も更新できなかった」を repository が domain.ErrNotFound で表す契約を handler 側から見る。
type fakeAiChatSessionRepo struct {
	session *domain.AiChatSession
	err     error
}

func (f *fakeAiChatSessionRepo) ListByUserID(context.Context, uint64) ([]domain.AiChatSession, error) {
	return nil, f.err
}

func (f *fakeAiChatSessionRepo) FindByID(context.Context, uint64) (*domain.AiChatSession, error) {
	if f.session == nil {
		return nil, domain.ErrNotFound
	}
	return f.session, nil
}

func (f *fakeAiChatSessionRepo) Create(context.Context, *domain.AiChatSession) error { return f.err }

func (f *fakeAiChatSessionRepo) UpdateTitle(context.Context, uint64, string) error { return f.err }

func (f *fakeAiChatSessionRepo) Delete(context.Context, uint64) error { return f.err }

func newAiChatHandlerForTitle(repo repository.AiChatSessionRepository) *AiChatHandler {
	return NewAiChatHandler(
		usecase.NewGetAiChatSessionsByUserIDUseCase(repo),
		nil, // 作成 usecase はこのテストで使わない
		usecase.NewGetAiChatSessionUseCase(repo),
		usecase.NewUpdateAiChatSessionTitleUseCase(repo),
		usecase.NewDeleteAiChatSessionUseCase(repo),
		nil,
	)
}

// 0 行更新（そのセッションが無い）は repository が domain.ErrNotFound を返す。
// 以前は 200 を返したうえに、直後の読み直しも失敗して本文 null の 200 になっていた
// （保存されていないタイトルを保存済みとして返していた）。
func Test_AIチャットハンドラ_セッションタイトル更新_対象なしは404(t *testing.T) {
	h := newAiChatHandlerForTitle(&fakeAiChatSessionRepo{err: domain.ErrNotFound})
	w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 7, "5")
	h.UpdateSessionTitle(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッションタイトル更新_成功時は200(t *testing.T) {
	h := newAiChatHandlerForTitle(&fakeAiChatSessionRepo{
		session: &domain.AiChatSession{ID: 5, UserID: 7, Title: "X"},
	})
	w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 7, "5")
	h.UpdateSessionTitle(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}
