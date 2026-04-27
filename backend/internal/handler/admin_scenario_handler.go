package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type AdminScenarioHandler struct {
	list   *usecase.ListAdminScenariosUseCase
	create *usecase.CreateAdminScenarioUseCase
	update *usecase.UpdateAdminScenarioUseCase
	del    *usecase.DeleteAdminScenarioUseCase
}

func NewAdminScenarioHandler(l *usecase.ListAdminScenariosUseCase, c *usecase.CreateAdminScenarioUseCase, u *usecase.UpdateAdminScenarioUseCase, d *usecase.DeleteAdminScenarioUseCase) *AdminScenarioHandler {
	return &AdminScenarioHandler{list: l, create: c, update: u, del: d}
}

func (h *AdminScenarioHandler) List(c *gin.Context) {
	rows, err := h.list.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *AdminScenarioHandler) Create(c *gin.Context) {
	var s domain.PracticeScenario
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), &s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *AdminScenarioHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var s domain.PracticeScenario
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.ID = id
	got, err := h.update.Execute(c.Request.Context(), &s)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, got)
}

func (h *AdminScenarioHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.del.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
