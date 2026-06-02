package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type UserStatsHandler struct {
	get *usecase.GetUserStatsUseCase
}

func NewUserStatsHandler(g *usecase.GetUserStatsUseCase) *UserStatsHandler {
	return &UserStatsHandler{get: g}
}

var (
	errUserStatsForbidden    = errors.New("forbidden")
	errUserStatsUnauthorized = errors.New("unauthorized")
)

// resolveUserID は "me" / 空文字を current user に、数字は current user 一致時のみ通す（IDOR 対策）。
func (h *UserStatsHandler) resolveUserID(c *gin.Context) (uint64, error) {
	cur := middleware.CurrentUserIDOrZero(c)
	if cur == 0 {
		return 0, errUserStatsUnauthorized
	}
	param := c.Param("userId")
	if param == "" || param == "me" {
		return cur, nil
	}
	uid, err := strconv.ParseUint(param, 10, 64)
	if err != nil {
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errUserStatsForbidden
	}
	return uid, nil
}

// @Summary      ユーザー 統計 取得
// @Description  指定 user (or 'me') の マイページ 集計 (合計 セッション / 平均 スコア)。 他 user 指定 は 403。
// @Tags         profile
// @Produce      json
// @Param        userId  path      string  true  "数字 ID または 'me'"
// @Success      200     {object}  github_com_norman6464_FreStyle_backend_internal_domain.UserStats
// @Failure      401     {object}  errorResponse  "未 認証"
// @Failure      403     {object}  errorResponse  "他 user 指定"
// @Failure      400     {object}  errorResponse  "DB 失敗"
// @Router       /profile/{userId}/stats [get]
// @Security     CookieAuth
func (h *UserStatsHandler) Get(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		switch err {
		case errUserStatsUnauthorized:
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errUserStatsForbidden:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	stats, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
