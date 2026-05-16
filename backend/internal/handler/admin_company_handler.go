package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type AdminCompanyHandler struct {
	list *usecase.ListCompaniesUseCase
}

func NewAdminCompanyHandler(l *usecase.ListCompaniesUseCase) *AdminCompanyHandler {
	return &AdminCompanyHandler{list: l}
}

// @Summary      会社 一覧 (SuperAdmin)
// @Description  全 company を 返す。 super_admin 専用 画面 用。 認可 は middleware で 別途 担保。
// @Tags         admin
// @Produce      json
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.Company
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /admin/companies [get]
// @Security     CookieAuth
func (h *AdminCompanyHandler) List(c *gin.Context) {
	companies, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, companies)
}
