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
	"gorm.io/gorm"
)

func init() { gin.SetMode(gin.TestMode) }

func TestActorFromContext(t *testing.T) {
	t.Run("認証済み user から id/companyID/role を取り出す", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		companyID := uint64(7)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 42, CompanyID: &companyID, Role: domain.RoleCompanyAdmin})

		uid, cid, role, ok := actorFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, uint64(42), uid)
		assert.Equal(t, uint64(7), cid)
		assert.Equal(t, domain.RoleCompanyAdmin, role)
	})

	t.Run("会社未所属(nil)なら companyID は 0", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(middleware.ContextKeyCurrentUser, &domain.User{ID: 1, CompanyID: nil, Role: domain.RoleSuperAdmin})

		_, cid, _, ok := actorFromContext(c)

		assert.True(t, ok)
		assert.Equal(t, uint64(0), cid)
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
		{"レコード未検出は 404", gorm.ErrRecordNotFound, http.StatusNotFound},
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

func TestUserCompanyIDValue(t *testing.T) {
	cid := uint64(9)
	assert.Equal(t, uint64(9), domain.User{CompanyID: &cid}.CompanyIDValue())
	assert.Equal(t, uint64(0), domain.User{CompanyID: nil}.CompanyIDValue())
}
