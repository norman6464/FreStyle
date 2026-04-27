package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ChatHandler struct {
	getRooms   *usecase.GetChatRoomsByUserIDUseCase
	createRoom *usecase.CreateChatRoomUseCase
}

func NewChatHandler(g *usecase.GetChatRoomsByUserIDUseCase, c *usecase.CreateChatRoomUseCase) *ChatHandler {
	return &ChatHandler{getRooms: g, createRoom: c}
}

func (h *ChatHandler) GetRooms(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.getRooms.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createRoomReq struct {
	Name    string `json:"name" binding:"required"`
	IsGroup bool   `json:"isGroup"`
}

func (h *ChatHandler) CreateRoom(c *gin.Context) {
	var req createRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.createRoom.Execute(c.Request.Context(), req.Name, req.IsGroup)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}
