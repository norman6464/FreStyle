package handler

import (
	"net/http"
	"testing"
)

func TestSessionNoteHandler_Upsert_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestSessionNoteHandler_Upsert_BadJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
