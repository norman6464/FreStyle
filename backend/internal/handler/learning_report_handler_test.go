package handler

import (
	"net/http"
	"testing"
)

func TestLearningReportHandler_List_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodGet, "", 0, "")
	(&LearningReportHandler{}).List(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestLearningReportHandler_Request_Unauthorized(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `{}`, 0, "")
	(&LearningReportHandler{}).Request(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestLearningReportHandler_Request_BadJSON(t *testing.T) {
	w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
	(&LearningReportHandler{}).Request(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}
