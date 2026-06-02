package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// HealthHandler は /api/v2/health エンドポイントを提供する。
type HealthHandler struct {
	uc *usecase.CheckHealthUseCase
}

func NewHealthHandler(uc *usecase.CheckHealthUseCase) *HealthHandler {
	return &HealthHandler{uc: uc}
}

// Get は DB 疎通を確認し UP / DOWN を返す。
//
//	@Summary      ヘルスチェック
//	@Description  バックエンド と DB の 疎通 を 確認 する。 ALB / CloudWatch / 監視 から 叩く 想定。
//	@Tags         health
//	@Produce      json
//	@Success      200  {object}  domain.Health  "UP"
//	@Failure      503  {object}  domain.Health  "DB 切断 等 で DOWN"
//	@Router       /health [get]
func (h *HealthHandler) Get(c *gin.Context) {
	result := h.uc.Execute(c.Request.Context())
	status := http.StatusOK
	if result.Status == domain.StatusDown {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, result)
}
