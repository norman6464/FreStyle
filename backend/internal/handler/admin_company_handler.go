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

// List は GET /api/v2/admin/companies。super_admin のみアクセス可。
func (h *AdminCompanyHandler) List(c *gin.Context) {
	companies, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, companies)
}
