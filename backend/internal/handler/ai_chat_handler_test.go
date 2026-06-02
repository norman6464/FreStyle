package handler

import (
	"net/http"
	"testing"
)

// ai_chat_handler のガード分岐（401 / 400）を zero-value handler で検証する。
// いずれも usecase 到達前に早期 return するため nil usecase で安全。

func TestAiChatHandler_GetSessions_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "")
	(&AiChatHandler{}).GetSessions(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAiChatHandler_CreateSession_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&AiChatHandler{}).CreateSession(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAiChatHandler_CreateSession_BadJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&AiChatHandler{}).CreateSession(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAiChatHandler_GetSession_BadID(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "abc")
	(&AiChatHandler{}).GetSession(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAiChatHandler_UpdateSessionTitle_BadID(t *testing.T) {
	w, c := noteCtx(http.MethodPut, `{"title":"X"}`, 0, "abc")
	(&AiChatHandler{}).UpdateSessionTitle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAiChatHandler_UpdateSessionTitle_MissingTitle(t *testing.T) {
	w, c := noteCtx(http.MethodPut, `{}`, 0, "5")
	(&AiChatHandler{}).UpdateSessionTitle(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAiChatHandler_DeleteSession_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodDelete, "", 0, "5")
	(&AiChatHandler{}).DeleteSession(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}
