package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AdminAuditHandler は監査ログの閲覧（super_admin 専用）を扱う。
type AdminAuditHandler struct {
	list *usecase.ListAuditEventsUseCase
}

func NewAdminAuditHandler(l *usecase.ListAuditEventsUseCase) *AdminAuditHandler {
	return &AdminAuditHandler{list: l}
}

// List は監査ログを新しい順で返す（super_admin 専用）。
//
//	@Summary      監査ログ一覧（super_admin）
//	@Description  管理者の重要操作（会社の有効/無効・従業員の停止/削除・招待など）の監査記録を新しい順で返す。super_admin 専用。
//	@Tags         admin
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.AuditEvent
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      403  {object}  errorResponse  "super_admin 以外"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /admin/audit-events [get]
//	@Security     CookieAuth
func (h *AdminAuditHandler) List(c *gin.Context) {
	if !isSuperAdmin(middleware.CurrentUserFromContext(c)) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
