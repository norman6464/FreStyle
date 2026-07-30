package handler

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type fakeNoteImagePresigner struct {
	url *domain.NoteImageUploadURL
	err error
}

func (f fakeNoteImagePresigner) Generate(context.Context, uint64, string, int64) (*domain.NoteImageUploadURL, error) {
	return f.url, f.err
}

func newNoteImageHandler(p repository.NoteImagePresigner) *NoteImageHandler {
	return NewNoteImageHandler(usecase.NewIssueNoteImageUploadURLUseCase(p))
}

func Test_ノート画像ハンドラ_アップロードURL発行(t *testing.T) {
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
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png","sizeBytes":1024}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("contentType 無し → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"sizeBytes":1024}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("sizeBytes 無し → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png"}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("許可外 MIME → 415", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"text/html","sizeBytes":1024}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("want 415, got %d", w.Code)
		}
	})
	t.Run("サイズ上限超過 → 413", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png","sizeBytes":5242881}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("want 413, got %d", w.Code)
		}
	})
	t.Run("空 body → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, ``, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("presigner エラー → 500（内部エラーはクライアントエラーにしない）", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png","sizeBytes":1024}`, 7, "")
		newNoteImageHandler(fakeNoteImagePresigner{err: context.DeadlineExceeded}).IssueUploadURL(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("want 500, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), "context deadline exceeded") {
			t.Fatalf("内部エラーの詳細が漏れている: %s", w.Body.String())
		}
	})
}
