package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// HealthHandler は /api/v2/health エンドポイントを提供する。
// Spring Boot の /actuator/health と並行運用する。
type HealthHandler struct {
	uc *usecase.CheckHealthUseCase
}

func NewHealthHandler(uc *usecase.CheckHealthUseCase) *HealthHandler {
	return &HealthHandler{uc: uc}
}

func (h *HealthHandler) Get(c *gin.Context) {
	result := h.uc.Execute(c.Request.Context())
	status := http.StatusOK
	if result.Status == domain.StatusDown {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, result)
}
