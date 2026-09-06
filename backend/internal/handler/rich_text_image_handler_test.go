package handler

import (
	"context"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type fakeRichTextImagePresigner struct {
	url *domain.RichTextImageUploadURL
	err error
}

func (f fakeRichTextImagePresigner) Generate(context.Context, uint64, string) (*domain.RichTextImageUploadURL, error) {
	return f.url, f.err
}

func newRichTextImageHandler(p repository.RichTextImagePresigner) *RichTextImageHandler {
	return NewRichTextImageHandler(usecase.NewIssueRichTextImageUploadURLUseCase(p))
}

func Test_リッチテキスト画像ハンドラ_アップロードURL発行(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `{}`, 0, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正な JSON → 400", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `not-json`, 7, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `{"contentType":"image/png"}`, 7, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{url: &domain.RichTextImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("presigner エラー → 400", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `{"contentType":"image/png"}`, 7, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{err: context.DeadlineExceeded}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
