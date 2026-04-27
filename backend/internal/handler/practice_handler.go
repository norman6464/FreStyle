package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type PracticeHandler struct {
	list *usecase.ListPracticeScenariosUseCase
	get  *usecase.GetPracticeScenarioUseCase
}

func NewPracticeHandler(l *usecase.ListPracticeScenariosUseCase, g *usecase.GetPracticeScenarioUseCase) *PracticeHandler {
	return &PracticeHandler{list: l, get: g}
}

func (h *PracticeHandler) List(c *gin.Context) {
	rows, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *PracticeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	s, err := h.get.Execute(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if s == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	c.JSON(http.StatusOK, s)
}
