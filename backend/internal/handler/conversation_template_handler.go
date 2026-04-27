package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type ConversationTemplateHandler struct {
	list *usecase.ListConversationTemplatesUseCase
}

func NewConversationTemplateHandler(l *usecase.ListConversationTemplatesUseCase) *ConversationTemplateHandler {
	return &ConversationTemplateHandler{list: l}
}

func (h *ConversationTemplateHandler) List(c *gin.Context) {
	rows, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}
