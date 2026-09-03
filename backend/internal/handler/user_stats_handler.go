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
		//nolint:nilerr // 数字以外の userId は current user にフォールバックする設計（err は握り潰さず意図的に無視）
		return cur, nil
	}
	if uid == 0 || uid != cur {
		return 0, errUserStatsForbidden
	}
	return uid, nil
}

func (h *UserStatsHandler) Get(c *gin.Context) {
	uid, err := h.resolveUserID(c)
	if err != nil {
		switch {
		case errors.Is(err, errUserStatsUnauthorized):
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		case errors.Is(err, errUserStatsForbidden):
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
