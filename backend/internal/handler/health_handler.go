package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/health"
)

// HealthHandler は /api/v2/health エンドポイントを提供する。
type HealthHandler struct {
	uc *health.CheckHealthUseCase
}

func NewHealthHandler(uc *health.CheckHealthUseCase) *HealthHandler {
	return &HealthHandler{uc: uc}
}

// Get は DB 疎通を確認し UP / DOWN を返す。
func (h *HealthHandler) Get(c *gin.Context) {
	result := h.uc.Execute(c.Request.Context())
	status := http.StatusOK
	if result.Status == domain.StatusDown {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, result)
}
