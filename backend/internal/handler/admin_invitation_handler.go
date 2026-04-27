package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

type AdminInvitationHandler struct {
	list   *usecase.ListAdminInvitationsUseCase
	create *usecase.CreateAdminInvitationUseCase
	cancel *usecase.CancelAdminInvitationUseCase
}

func NewAdminInvitationHandler(l *usecase.ListAdminInvitationsUseCase, c *usecase.CreateAdminInvitationUseCase, x *usecase.CancelAdminInvitationUseCase) *AdminInvitationHandler {
	return &AdminInvitationHandler{list: l, create: c, cancel: x}
}

func (h *AdminInvitationHandler) List(c *gin.Context) {
	cid, _ := strconv.ParseUint(c.Query("companyId"), 10, 64)
	rows, err := h.list.Execute(c.Request.Context(), cid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

type createAdminInvReq struct {
	CompanyID   uint64 `json:"companyId" binding:"required"`
	Email       string `json:"email" binding:"required"`
	Role        string `json:"role" binding:"required"`
	DisplayName string `json:"displayName"`
}

func (h *AdminInvitationHandler) Create(c *gin.Context) {
	var req createAdminInvReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	got, err := h.create.Execute(c.Request.Context(), usecase.CreateAdminInvitationInput{
		CompanyID: req.CompanyID, Email: req.Email, Role: req.Role, DisplayName: req.DisplayName,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, got)
}

func (h *AdminInvitationHandler) Cancel(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.cancel.Execute(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
