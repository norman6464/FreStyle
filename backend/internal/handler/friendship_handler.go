package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type FriendshipHandler struct {
	list    *usecase.ListFriendshipsUseCase
	request *usecase.RequestFriendshipUseCase
	respond *usecase.RespondFriendshipUseCase
}

func NewFriendshipHandler(l *usecase.ListFriendshipsUseCase, r *usecase.RequestFriendshipUseCase, p *usecase.RespondFriendshipUseCase) *FriendshipHandler {
	return &FriendshipHandler{list: l, request: r, respond: p}
}

func (h *FriendshipHandler) List(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type requestFriendshipReq struct {
	RequesterID uint64 `json:"requesterId" binding:"required"`
	AddresseeID uint64 `json:"addresseeId" binding:"required"`
}

func (h *FriendshipHandler) Request(c *gin.Context) {
	var req requestFriendshipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.request.Execute(c.Request.Context(), req.RequesterID, req.AddresseeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

type respondFriendshipReq struct {
	Accept bool `json:"accept"`
}

func (h *FriendshipHandler) Respond(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req respondFriendshipReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.respond.Execute(c.Request.Context(), id, req.Accept); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
