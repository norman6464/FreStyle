package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type NoteHandler struct {
	list   *usecase.ListNotesByUserIDUseCase
	create *usecase.CreateNoteUseCase
	update *usecase.UpdateNoteUseCase
	del    *usecase.DeleteNoteUseCase
}

func NewNoteHandler(
	l *usecase.ListNotesByUserIDUseCase,
	c *usecase.CreateNoteUseCase,
	u *usecase.UpdateNoteUseCase,
	d *usecase.DeleteNoteUseCase,
) *NoteHandler {
	return &NoteHandler{list: l, create: c, update: u, del: d}
}

func (h *NoteHandler) List(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type noteCreateReq struct {
	UserID   uint64 `json:"userId" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	IsPublic bool   `json:"isPublic"`
}

func (h *NoteHandler) Create(c *gin.Context) {
	var req noteCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateNoteInput{
		UserID: req.UserID, Title: req.Title, Content: req.Content, IsPublic: req.IsPublic,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

type noteUpdateReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	IsPublic bool   `json:"isPublic"`
}

func (h *NoteHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req noteUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.update.Execute(c.Request.Context(), usecase.UpdateNoteInput{
		ID: id, Title: req.Title, Content: req.Content, IsPublic: req.IsPublic,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}

func (h *NoteHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.del.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
