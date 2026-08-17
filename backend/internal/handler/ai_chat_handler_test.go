package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// fakeAiChatSessionRepo は AiChatSessionRepository の handler テスト用 fake。
type fakeAiChatSessionRepo struct {
	found   *domain.AiChatSession
	findErr error
}

func (f *fakeAiChatSessionRepo) ListByUserID(_ context.Context, _ uint64) ([]domain.AiChatSession, error) {
	return nil, nil
}
func (f *fakeAiChatSessionRepo) FindByID(_ context.Context, _ uint64) (*domain.AiChatSession, error) {
	return f.found, f.findErr
}
func (f *fakeAiChatSessionRepo) Create(_ context.Context, _ *domain.AiChatSession) error { return nil }
func (f *fakeAiChatSessionRepo) UpdateTitle(_ context.Context, _ uint64, _ string) error { return nil }
func (f *fakeAiChatSessionRepo) Delete(_ context.Context, _ uint64) error                { return nil }

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
	w, c := noteCtx(http.MethodGet, "", 7, "abc")
	(&AiChatHandler{}).GetSession(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション取得_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "5")
	(&AiChatHandler{}).GetSession(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション取得_他人のセッションは403(t *testing.T) {
	uc := usecase.NewGetAiChatSessionUseCase(&fakeAiChatSessionRepo{found: &domain.AiChatSession{ID: 5, UserID: 99}})
	w, c := noteCtx(http.MethodGet, "", 7, "5")
	(&AiChatHandler{getSession: uc}).GetSession(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション取得_不存在は404(t *testing.T) {
	uc := usecase.NewGetAiChatSessionUseCase(&fakeAiChatSessionRepo{findErr: gorm.ErrRecordNotFound})
	w, c := noteCtx(http.MethodGet, "", 7, "5")
	(&AiChatHandler{getSession: uc}).GetSession(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func Test_AIチャットハンドラ_セッション取得_本人は200(t *testing.T) {
	uc := usecase.NewGetAiChatSessionUseCase(&fakeAiChatSessionRepo{found: &domain.AiChatSession{ID: 5, UserID: 7}})
	w, c := noteCtx(http.MethodGet, "", 7, "5")
	(&AiChatHandler{getSession: uc}).GetSession(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
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
