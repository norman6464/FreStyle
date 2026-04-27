package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type FavoritePhraseHandler struct {
	list   *usecase.ListFavoritePhrasesUseCase
	add    *usecase.AddFavoritePhraseUseCase
	remove *usecase.DeleteFavoritePhraseUseCase
}

func NewFavoritePhraseHandler(l *usecase.ListFavoritePhrasesUseCase, a *usecase.AddFavoritePhraseUseCase, d *usecase.DeleteFavoritePhraseUseCase) *FavoritePhraseHandler {
	return &FavoritePhraseHandler{list: l, add: a, remove: d}
}

func (h *FavoritePhraseHandler) List(c *gin.Context) {
	uid, _ := strconv.ParseUint(c.Query("userId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type addPhraseReq struct {
	UserID uint64 `json:"userId" binding:"required"`
	Phrase string `json:"phrase" binding:"required"`
	Note   string `json:"note"`
}

func (h *FavoritePhraseHandler) Add(c *gin.Context) {
	var req addPhraseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.add.Execute(c.Request.Context(), req.UserID, req.Phrase, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *FavoritePhraseHandler) Remove(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.remove.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
