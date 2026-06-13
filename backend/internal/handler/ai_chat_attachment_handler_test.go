package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type fakeAttachPresigner struct {
	url *repository.AiChatAttachmentUploadURL
	err error
}

func (f fakeAttachPresigner) Generate(context.Context, uint64, string, string) (*repository.AiChatAttachmentUploadURL, error) {
	return f.url, f.err
}

func newAttachHandler(p repository.AiChatAttachmentPresigner) *AiChatAttachmentHandler {
	return NewAiChatAttachmentHandler(usecase.NewIssueAiChatAttachmentUploadURLUseCase(p))
}

func TestAiChatAttachmentHandler_IssueUploadURL(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{}`, 0, "")
		newAttachHandler(fakeAttachPresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("usecase 未設定 -> 503", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{}`, 7, "")
		(&AiChatAttachmentHandler{}).IssueUploadURL(c)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", w.Code)
		}
	})
	t.Run("不正な JSON → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
		newAttachHandler(fakeAttachPresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("contentType 欠落 → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"filename":"a.png","sizeBytes":100}`, 7, "")
		newAttachHandler(fakeAttachPresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
