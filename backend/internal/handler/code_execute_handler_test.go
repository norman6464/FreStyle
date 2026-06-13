package handler

import (
	"net/http"
	"testing"
)

func Test_コード実行ハンドラ_実行_不正なJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 0, "")
	(&CodeExecuteHandler{}).Execute(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
