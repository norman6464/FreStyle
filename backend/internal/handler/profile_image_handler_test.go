package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

type fakeProfileImagePresigner struct {
	url *domain.ProfileImageUploadURL
	err error
}

func (f fakeProfileImagePresigner) Generate(context.Context, uint64, string, string) (*domain.ProfileImageUploadURL, error) {
	return f.url, f.err
}

func newProfileImageHandler(p repository.ProfileImagePresigner) *ProfileImageHandler {
	return NewProfileImageHandler(usecase.NewIssueProfileImageUploadURLUseCase(p))
}

// userIDCtx は cur user（ContextKeyCurrentUserID）+ userId param + JSON body を持つ context。
func userIDCtx(body string, cur uint64, userIDParam string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if cur != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, cur)
	}
	c.Params = gin.Params{{Key: "userId", Value: userIDParam}}
	return w, c
}

func TestProfileImageHandler_IssueUploadURL(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := userIDCtx(`{}`, 0, "me")
		newProfileImageHandler(fakeProfileImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("forbidden (他人指定)", func(t *testing.T) {
		w, c := userIDCtx(`{}`, 7, "99")
		newProfileImageHandler(fakeProfileImagePresigner{}).IssueUploadURL(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("want 403, got %d", w.Code)
		}
	})
	t.Run("presigner エラー → 400", func(t *testing.T) {
		w, c := userIDCtx(`{"contentType":"image/png"}`, 7, "me")
		newProfileImageHandler(fakeProfileImagePresigner{err: context.DeadlineExceeded}).IssueUploadURL(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := userIDCtx(`{"contentType":"image/png"}`, 7, "me")
		newProfileImageHandler(fakeProfileImagePresigner{url: &domain.ProfileImageUploadURL{}}).IssueUploadURL(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
}
