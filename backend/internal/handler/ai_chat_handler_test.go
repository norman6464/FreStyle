package handler

import (
	"net/http"
	"testing"
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
