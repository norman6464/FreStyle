package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
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

// List は current user の note 一覧を返す。
// userId はクライアントから受け取らない（IDOR 対策）。
func (h *NoteHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type noteCreateReq struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content"`
	IsPublic bool   `json:"isPublic"`
}

// Create は current user 名義で note を作る。userId は受け取らず固定する。
func (h *NoteHandler) Create(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req noteCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateNoteInput{
		UserID: uid, Title: req.Title, Content: req.Content, IsPublic: req.IsPublic,
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

// Update は current user 所有の note のみ更新可能。usecase 側で所有者検証する。
func (h *NoteHandler) Update(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req noteUpdateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.update.Execute(c.Request.Context(), usecase.UpdateNoteInput{
		UserID: uid, ID: id, Title: req.Title, Content: req.Content, IsPublic: req.IsPublic,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrNoteForbidden) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}

// Delete は WHERE で user_id を絞るため他人の note は消せない。
func (h *NoteHandler) Delete(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.del.Execute(c.Request.Context(), uid, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
