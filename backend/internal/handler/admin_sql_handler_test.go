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

// fakeSQLConsoleRepo は handler テスト用の最小 fake。
type fakeSQLConsoleRepo struct {
	called bool
	result *repository.SQLQueryResult
}

func (f *fakeSQLConsoleRepo) RunReadOnly(context.Context, string, int) (*repository.SQLQueryResult, error) {
	f.called = true
	if f.result != nil {
		return f.result, nil
	}
	return &repository.SQLQueryResult{Columns: []string{"n"}, Rows: [][]any{{int64(1)}}}, nil
}

func newAdminSQLHandler(repo *fakeSQLConsoleRepo) *AdminSQLHandler {
	return NewAdminSQLHandler(usecase.NewExecuteReadOnlySQLUseCase(repo))
}

func postSQL(t *testing.T, actor *domain.User, body string, repo *fakeSQLConsoleRepo) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/admin/sql", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	if actor != nil {
		c.Set(middleware.ContextKeyCurrentUser, actor)
	}
	newAdminSQLHandler(repo).Run(c)
	return w
}

func TestAdminSQLHandler_SuperAdmin_OK(t *testing.T) {
	repo := &fakeSQLConsoleRepo{}
	actor := &domain.User{ID: 1, Email: "ops@example.com", Role: domain.RoleSuperAdmin}

	w := postSQL(t, actor, `{"query":"SELECT 1 AS n"}`, repo)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d (%s)", w.Code, w.Body.String())
	}
	if !repo.called {
		t.Fatal("repository が呼ばれていない")
	}
}

func TestAdminSQLHandler_NonSuperAdmin_Forbidden(t *testing.T) {
	for _, role := range []string{domain.RoleCompanyAdmin, domain.RoleTrainee} {
		t.Run(role, func(t *testing.T) {
			repo := &fakeSQLConsoleRepo{}
			actor := &domain.User{ID: 2, Email: "u@example.com", Role: role}

			w := postSQL(t, actor, `{"query":"SELECT 1"}`, repo)

			if w.Code != http.StatusForbidden {
				t.Fatalf("want 403, got %d", w.Code)
			}
			if repo.called {
				t.Fatal("非 super_admin で repository を呼んではならない")
			}
		})
	}
}

func TestAdminSQLHandler_NoUser_Forbidden(t *testing.T) {
	repo := &fakeSQLConsoleRepo{}
	w := postSQL(t, nil, `{"query":"SELECT 1"}`, repo)
	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d", w.Code)
	}
}

func TestAdminSQLHandler_MissingQuery_BadRequest(t *testing.T) {
	repo := &fakeSQLConsoleRepo{}
	actor := &domain.User{ID: 1, Email: "ops@example.com", Role: domain.RoleSuperAdmin}
	w := postSQL(t, actor, `{}`, repo)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestAdminSQLHandler_WriteQuery_BadRequest(t *testing.T) {
	repo := &fakeSQLConsoleRepo{}
	actor := &domain.User{ID: 1, Email: "ops@example.com", Role: domain.RoleSuperAdmin}

	w := postSQL(t, actor, `{"query":"DELETE FROM users"}`, repo)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
	if repo.called {
		t.Fatal("write クエリで repository を呼んではならない")
	}
}
