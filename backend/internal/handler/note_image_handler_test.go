package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// テストで使う固定値。マジックナンバーを避けるため名前を付ける。
const (
	unauthenticatedUserID uint64 = 0
	testUserID            uint64 = 7
	validSizeBytes        int64  = 1024
	// overLimitSizeBytes は usecase の上限 5MB を 1 byte 超えた値。
	overLimitSizeBytes int64 = 5*1024*1024 + 1
)

// errBucketNotConfigured はタイムアウト以外（恒久的な設定不備）を表すテスト用エラー。
var errBucketNotConfigured = errors.New("s3: bucket name is required")

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

// uploadURLBody は contentType / sizeBytes を持つリクエスト body を組み立てる。
func uploadURLBody(contentType string, sizeBytes int64) string {
	return fmt.Sprintf(`{"contentType":%q,"sizeBytes":%d}`, contentType, sizeBytes)
}

func Test_ノート画像ハンドラ_アップロードURL発行(t *testing.T) {
	okPresigner := fakeNoteImagePresigner{url: &domain.NoteImageUploadURL{}}

	t.Run("未認証", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{}`, unauthenticatedUserID, "")
		newNoteImageHandler(fakeNoteImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("不正な JSON → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `not-json`, testUserID, "")
		newNoteImageHandler(fakeNoteImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, uploadURLBody("image/png", validSizeBytes), testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("contentType 無し → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, fmt.Sprintf(`{"sizeBytes":%d}`, validSizeBytes), testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("sizeBytes 無し → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, `{"contentType":"image/png"}`, testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("許可外 MIME → 415", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, uploadURLBody("text/html", validSizeBytes), testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("want 415, got %d", w.Code)
		}
	})
	t.Run("サイズ上限超過 → 413", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, uploadURLBody("image/png", overLimitSizeBytes), testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusRequestEntityTooLarge {
			t.Fatalf("want 413, got %d", w.Code)
		}
	})
	t.Run("空 body → 400", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, ``, testUserID, "")
		newNoteImageHandler(okPresigner).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("presigner タイムアウト → 503 + Retry-After（再試行で回復しうる一時障害）", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, uploadURLBody("image/png", validSizeBytes), testUserID, "")
		newNoteImageHandler(fakeNoteImagePresigner{err: context.DeadlineExceeded}).IssueUploadURL(c)
		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("want 503, got %d", w.Code)
		}
		if got := w.Header().Get("Retry-After"); got == "" {
			t.Fatal("Retry-After ヘッダが無い")
		}
		if strings.Contains(w.Body.String(), "context deadline exceeded") {
			t.Fatalf("内部エラーの詳細が漏れている: %s", w.Body.String())
		}
	})
	t.Run("presigner の恒久的エラー → 500（内部エラーはクライアントエラーにしない）", func(t *testing.T) {
		w, c := noteCtx(http.MethodPost, uploadURLBody("image/png", validSizeBytes), testUserID, "")
		newNoteImageHandler(fakeNoteImagePresigner{err: errBucketNotConfigured}).IssueUploadURL(c)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("want 500, got %d", w.Code)
		}
		if strings.Contains(w.Body.String(), errBucketNotConfigured.Error()) {
			t.Fatalf("内部エラーの詳細が漏れている: %s", w.Body.String())
		}
	})
}
