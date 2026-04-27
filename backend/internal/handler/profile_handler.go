package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ProfileHandler struct {
	get    *usecase.GetProfileUseCase
	update *usecase.UpdateProfileUseCase
}

func NewProfileHandler(g *usecase.GetProfileUseCase, u *usecase.UpdateProfileUseCase) *ProfileHandler {
	return &ProfileHandler{get: g, update: u}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	p, err := h.get.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "profile_not_found"})
		return
	}
	c.JSON(http.StatusOK, p)
}

type updateProfileReq struct {
	Bio       string `json:"bio"`
	AvatarURL string `json:"avatarUrl"`
}

func (h *ProfileHandler) Update(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	var req updateProfileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.update.Execute(c.Request.Context(), usecase.UpdateProfileInput{
		UserID: uid, Bio: req.Bio, AvatarURL: req.AvatarURL,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}
