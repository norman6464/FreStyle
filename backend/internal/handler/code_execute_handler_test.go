package handler

import (
	"net/http"
	"testing"
)

func TestCodeExecuteHandler_Execute_BadJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 0, "")
	(&CodeExecuteHandler{}).Execute(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
