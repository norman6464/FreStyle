package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() { gin.SetMode(gin.TestMode) }

// runWithAudit は audit middleware + ハンドラを 1 ルートで実行し、記録されたエントリを返す。
func runWithAudit(t *testing.T, actor *domain.User, status int) []middleware.AuditEntry {
	t.Helper()
	var got []middleware.AuditEntry
	r := gin.New()
	r.PATCH(
		"/admin/companies/:id/active",
		func(c *gin.Context) {
			if actor != nil {
				c.Set(middleware.ContextKeyCurrentUser, actor)
			}
		},
		middleware.AuditLog(func(_ context.Context, e middleware.AuditEntry) {
			got = append(got, e)
		}),
		func(c *gin.Context) { c.Status(status) },
	)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/admin/companies/7/active", nil)
	r.ServeHTTP(w, req)
	require.Equal(t, status, w.Code)
	return got
}

func Test_監査middleware_成功した変更操作を記録する(t *testing.T) {
	got := runWithAudit(t, &domain.User{ID: 9, Email: "a@x", Role: domain.RoleSuperAdmin}, http.StatusOK)
	require.Len(t, got, 1)
	assert.Equal(t, uint64(9), got[0].ActorID)
	assert.Equal(t, "a@x", got[0].ActorEmail)
	assert.Equal(t, "PATCH /admin/companies/:id/active", got[0].Action)
	assert.Equal(t, uint64(7), got[0].TargetID)
}

func Test_監査middleware_失敗レスポンスは記録しない(t *testing.T) {
	got := runWithAudit(t, &domain.User{ID: 9, Role: domain.RoleSuperAdmin}, http.StatusForbidden)
	assert.Empty(t, got)
}

func Test_監査middleware_actorなしは記録しない(t *testing.T) {
	got := runWithAudit(t, nil, http.StatusOK)
	assert.Empty(t, got)
}
