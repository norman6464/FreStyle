package handler

import (
	"net/http"
	"testing"
)

func Test_学習レポートハンドラ_一覧_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "")
	(&LearningReportHandler{}).List(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_学習レポートハンドラ_要求_未認証(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&LearningReportHandler{}).Request(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func Test_学習レポートハンドラ_要求_不正なJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&LearningReportHandler{}).Request(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
