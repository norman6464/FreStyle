package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type fakeNoteImagePresigner struct {
	url *domain.NoteImageUploadURL
	err error
}

func (f fakeNoteImagePresigner) Generate(context.Context, uint64, string) (*domain.NoteImageUploadURL, error) {
	return f.url, f.err
}

func newNoteImageHandler(p repository.NoteImagePresigner) *NoteImageHandler {
	return NewNoteImageHandler(usecase.NewIssueNoteImageUploadURLUseCase(p))
}

func TestNoteImageHandler_IssueUploadURL(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{}`, 0, "")
		newNoteImageHandler(fakeNoteImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正な JSON → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `not-json`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png"}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("presigner エラー → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png"}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{err: context.DeadlineExceeded}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
