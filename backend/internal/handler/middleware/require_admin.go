package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// RequireAdmin は super_admin / company_admin 以外を /admin/* 系から締め出す。
//
// 各 handler にも role 検査はあるが、それだけだと 1 箇所書き忘れただけで穴になる
// （実際 GET /admin/companies で認可が抜けており、trainee が全顧客企業を列挙できた:
// FRESTYLE-76。招待取消でも同種の漏れがあった: FRESTYLE-228）。
// 入口で非管理者を落とす多層防御として置く。
//
// super_admin 限定か company_admin も可かの細かい判定は、引き続き各 handler / usecase
// が行う（admin 配下には company_admin が使う会員管理・招待も含まれるため、
// ここで super_admin 限定にはできない）。
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := CurrentUserFromContext(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if user.Role != domain.RoleSuperAdmin && user.Role != domain.RoleCompanyAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
