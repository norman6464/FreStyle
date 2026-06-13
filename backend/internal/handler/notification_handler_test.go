package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// fakeNotifRepo は repository.NotificationRepository の最小 fake。
type fakeNotifRepo struct {
	rows  []domain.Notification
	count int64
	err   error
}

func (f *fakeNotifRepo) Create(context.Context, *domain.Notification) error { return f.err }
func (f *fakeNotifRepo) ListByUserID(context.Context, uint64) ([]domain.Notification, error) {
	return f.rows, f.err
}
func (f *fakeNotifRepo) MarkRead(context.Context, uint64, uint64) error { return f.err }
func (f *fakeNotifRepo) MarkAllRead(context.Context, uint64) error      { return f.err }
func (f *fakeNotifRepo) CountUnread(context.Context, uint64) (int64, error) {
	return f.count, f.err
}

func newNotifHandler(repo repository.NotificationRepository) *NotificationHandler {
	return NewNotificationHandler(
		usecase.NewListNotificationsUseCase(repo),
		usecase.NewMarkNotificationReadUseCase(repo),
		usecase.NewMarkAllNotificationsReadUseCase(repo),
		usecase.NewCountUnreadNotificationsUseCase(repo),
	)
}

func notifCtx(uid uint64, idVal string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if uid != 0 {
		c.Set(middleware.ContextKeyCurrentUserID, uid)
	}
	if idVal != "" {
		c.Params = gin.Params{{Key: "id", Value: idVal}}
	}
	return w, c
}

func Test_通知ハンドラ_一覧(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := notifCtx(0, "")
		newNotifHandler(&fakeNotifRepo{}).List(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{rows: []domain.Notification{{ID: 1}}}).List(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{err: context.DeadlineExceeded}).List(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_通知ハンドラ_既読化(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := notifCtx(0, "1")
		newNotifHandler(&fakeNotifRepo{}).MarkRead(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系 → 204", func(t *testing.T) {
		_, c := notifCtx(7, "1")
		newNotifHandler(&fakeNotifRepo{}).MarkRead(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := notifCtx(7, "1")
		newNotifHandler(&fakeNotifRepo{err: context.DeadlineExceeded}).MarkRead(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_通知ハンドラ_全既読化(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := notifCtx(0, "")
		newNotifHandler(&fakeNotifRepo{}).MarkAllRead(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系 → 204", func(t *testing.T) {
		_, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{}).MarkAllRead(c)
		if c.Writer.Status() != http.StatusNoContent {
			t.Fatalf("want 204, got %d", c.Writer.Status())
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{err: context.DeadlineExceeded}).MarkAllRead(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}

func Test_通知ハンドラ_未読数(t *testing.T) {
	t.Run("未認証", func(t *testing.T) {
		w, c := notifCtx(0, "")
		newNotifHandler(&fakeNotifRepo{}).UnreadCount(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("want 401, got %d", w.Code)
		}
	})
	t.Run("正常系", func(t *testing.T) {
		w, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{count: 3}).UnreadCount(c)
		if w.Code != http.StatusOK {
			t.Fatalf("want 200, got %d", w.Code)
		}
	})
	t.Run("リポジトリエラー → 400", func(t *testing.T) {
		w, c := notifCtx(7, "")
		newNotifHandler(&fakeNotifRepo{err: context.DeadlineExceeded}).UnreadCount(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("want 400, got %d", w.Code)
		}
	})
}
