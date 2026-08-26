package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/stretchr/testify/assert"
)

func init() { gin.SetMode(gin.TestMode) }

func TestActorFromContext(t *testing.T) {
	t.Run("認証済み user から id/所属会社/role を取り出す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		companyID := uint64(7)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 42, CompanyID: &companyID, Role: domain.RoleCompanyAdmin})

		uid, company, role, ok := actorFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, uint64(42), uid)
		gotID, affiliated := company.CompanyID()
		assert.True(t, affiliated)
		assert.Equal(t, uint64(7), gotID)
		assert.Equal(t, domain.RoleCompanyAdmin, role)
	})

	t.Run("会社未所属(nil)なら未所属の CompanyRef", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, CompanyID: nil, Role: domain.RoleSuperAdmin})

		_, company, _, ok := actorFromContext(c)

		assert.True(t, ok)
		_, affiliated := company.CompanyID()
		assert.False(t, affiliated)
	})

	t.Run("未認証なら 401 を書き ok=false", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		_, _, _, ok := actorFromContext(c)

		assert.False(t, ok)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRespondEntityErr(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"レコード未検出は 404", domain.ErrNotFound, http.StatusNotFound},
		{"forbidden は 403", errors.New("forbidden"), http.StatusForbidden},
		{"forbidden 詳細付きも 403", errors.New("forbidden: only company_admin or super_admin can create materials"), http.StatusForbidden},
		{"会社未所属は 403", errors.New("actor must belong to a company"), http.StatusForbidden},
		{"その他は 500", errors.New("db down"), http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			respondEntityErr(c, tc.err, "見つかりません", "失敗しました")

			assert.Equal(t, tc.wantCode, w.Code)
		})
	}
}

func TestUserCompanyRef(t *testing.T) {
	cid := uint64(9)
	assert.Equal(t, domain.CompanyRefOf(9), domain.User{CompanyID: &cid}.CompanyRef())
	assert.Equal(t, domain.NoCompany(), domain.User{CompanyID: nil}.CompanyRef())
}
