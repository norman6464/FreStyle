package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase/richtextimage"
)

type fakeRichTextImagePresigner struct {
	url *domain.RichTextImageUploadURL
	err error
}

func (f fakeRichTextImagePresigner) Generate(context.Context, uint64, string, int64) (*domain.RichTextImageUploadURL, error) {
	return f.url, f.err
}

func newRichTextImageHandler(p repository.RichTextImagePresigner) *RichTextImageHandler {
	return NewRichTextImageHandler(richtextimage.NewIssueRichTextImageUploadURLUseCase(p))
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
	// FRESTYLE-9: Content-Type 許可リスト外・サイズ超過は専用のエラーコードで返す
	// （他の一般エラーと同じ "error" フィールドの形は保ったまま、値だけ判別できるようにする）。
	t.Run("許可リスト外のContentType → 400 unsupported_content_type", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `{"contentType":"image/svg+xml","size":1024}`, 7, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{err: domain.ErrUnsupportedImageContentType}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json body: %v", err)
		}
		if body["error"] != "unsupported_content_type" {
			t.Fatalf("want error=unsupported_content_type, got %+v", body)
		}
	})
	t.Run("サイズ超過 → 400 image_too_large", func(t *testing.T) {
		w, c := testCtx(http.MethodPost, `{"contentType":"image/png","size":99999999}`, 7, "")
		newRichTextImageHandler(fakeRichTextImagePresigner{err: domain.ErrImageTooLarge}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("invalid json body: %v", err)
		}
		if body["error"] != "image_too_large" {
			t.Fatalf("want error=image_too_large, got %+v", body)
		}
	})
}
