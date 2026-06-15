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
)

// fakeAuditRepoH は AuditRepository の最小 fake（handler テスト用）。
type fakeAuditRepoH struct {
	rows []domain.AuditEvent
}

func (f *fakeAuditRepoH) Record(context.Context, *domain.AuditEvent) error { return nil }
func (f *fakeAuditRepoH) ListRecent(context.Context, int) ([]domain.AuditEvent, error) {
	return f.rows, nil
}

func newAdminAuditHandler() *AdminAuditHandler {
	return NewAdminAuditHandler(usecase.NewListAuditEventsUseCase(&fakeAuditRepoH{
		rows: []domain.AuditEvent{{ID: 1, Action: "DELETE /admin/members/:userId"}},
	}))
}

func Test_監査ハンドラ_一覧_super_admin_正常系(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, Role: domain.RoleSuperAdmin})
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/audit-events", nil)
	newAdminAuditHandler().List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func Test_監査ハンドラ_一覧_super_admin以外は禁止(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 2, Role: domain.RoleTrainee})
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/audit-events", nil)
	newAdminAuditHandler().List(c)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}
