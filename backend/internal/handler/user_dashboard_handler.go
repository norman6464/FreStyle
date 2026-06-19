package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// UserDashboardHandler はパーソナライズダッシュボードデータを返す。
type UserDashboardHandler struct {
	getDashboard *usecase.GetUserDashboardUseCase
}

func NewUserDashboardHandler(d *usecase.GetUserDashboardUseCase) *UserDashboardHandler {
	return &UserDashboardHandler{getDashboard: d}
}

// Get はログイン中ユーザーのダッシュボードデータ（streak / 活動カレンダー / 「続きから」）を返す。
//
//	@Summary      ダッシュボード データ 取得
//	@Description  過去 90 日間の日次活動サマリー・連続学習日数・直近の閲覧章を返す。 カレンダーヒートマップ や 「続きから」 カード の 描画 に 使う。
//	@Tags         dashboard
//	@Produce      json
//	@Success      200  {object}  usecase.GetUserDashboardOutput
//	@Failure      401  {object}  errorResponse  "未認証"
//	@Failure      500  {object}  errorResponse  "集計失敗"
//	@Router       /me/dashboard [get]
//	@Security     CookieAuth
func (h *UserDashboardHandler) Get(c *gin.Context) {
	userID := middleware.CurrentUserIDOrZero(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	out, err := h.getDashboard.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ダッシュボードの取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, out)
}
