package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
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

// List は current user の favorite phrase 一覧を返す。
// userId はクライアントから受け取らない（IDOR 対策）。
func (h *FavoritePhraseHandler) List(c *gin.Context) {
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

// addPhraseReq は userId を受け取らない（current user 固定）。
// originalText / rephrasedText / pattern はフロントの言い換え保存 API に合わせるが、
// backend domain は phrase / note しか持たないため、phrase = rephrasedText, note = originalText で保存する。
type addPhraseReq struct {
	OriginalText  string `json:"originalText"`
	RephrasedText string `json:"rephrasedText"`
	Pattern       string `json:"pattern"`
	// 旧仕様の互換のため phrase/note も受け付ける。
	Phrase string `json:"phrase"`
	Note   string `json:"note"`
}

func (h *FavoritePhraseHandler) Add(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	var req addPhraseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	phrase := req.RephrasedText
	if phrase == "" {
		phrase = req.Phrase
	}
	note := req.OriginalText
	if note == "" {
		note = req.Note
	}
	if phrase == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phrase is required"})
		return
	}
	got, err := h.add.Execute(c.Request.Context(), uid, phrase, note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

// Remove は current user 所有の phrase のみ削除可能（IDOR 対策）。
func (h *FavoritePhraseHandler) Remove(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)
	if uid == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.remove.Execute(c.Request.Context(), uid, id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
