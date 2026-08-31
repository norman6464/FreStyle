package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name       string
		actor      *domain.User
		wantStatus int
		wantNext   bool
	}{
		{"super_admin は通す", &domain.User{ID: 1, Role: domain.RoleSuperAdmin}, http.StatusOK, true},
		{"company_admin は通す", &domain.User{ID: 2, Role: domain.RoleCompanyAdmin}, http.StatusOK, true},
		{"trainee は 403", &domain.User{ID: 3, Role: domain.RoleTrainee}, http.StatusForbidden, false},
		{"role 空は 403", &domain.User{ID: 4}, http.StatusForbidden, false},
		{"未認証は 401", nil, http.StatusUnauthorized, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false

			r := gin.New()
			// CurrentUser middleware の代わりに context へ actor を積む。
			r.Use(func(c *gin.Context) {
				if tc.actor != nil {
					c.Set(ContextKeyCurrentUser, tc.actor)
				}
				c.Next()
			})
			r.GET("/admin/members", RequireAdmin(), func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin/members", nil))

			if w.Code != tc.wantStatus {
				t.Fatalf("status: want %d, got %d", tc.wantStatus, w.Code)
			}
			if nextCalled != tc.wantNext {
				t.Fatalf("next 到達: want %v, got %v", tc.wantNext, nextCalled)
			}
		})
	}
}
