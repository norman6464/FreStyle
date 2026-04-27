package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type SharedSessionHandler struct {
	list   *usecase.ListSharedSessionsUseCase
	create *usecase.CreateSharedSessionUseCase
}

func NewSharedSessionHandler(l *usecase.ListSharedSessionsUseCase, c *usecase.CreateSharedSessionUseCase) *SharedSessionHandler {
	return &SharedSessionHandler{list: l, create: c}
}

func (h *SharedSessionHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.list.Execute(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createSharedReq struct {
	OwnerID     uint64 `json:"ownerId" binding:"required"`
	SessionID   uint64 `json:"sessionId" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"isPublic"`
}

func (h *SharedSessionHandler) Create(c *gin.Context) {
	var req createSharedReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateSharedSessionInput{
		OwnerID: req.OwnerID, SessionID: req.SessionID,
		Title: req.Title, Description: req.Description, IsPublic: req.IsPublic,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}
