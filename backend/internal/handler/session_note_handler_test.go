package handler

import (
	"net/http"
	"testing"
)

func Test_セッションノートハンドラ_保存_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_セッションノートハンドラ_保存_不正なJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&SessionNoteHandler{}).Upsert(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
