package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// TeachingMaterialHandler は教材 API を扱う。
//
//	GET    /api/v2/teaching-materials      一覧（actor role / company で自動フィルタ）
//	GET    /api/v2/teaching-materials/:id  詳細
//	POST   /api/v2/teaching-materials      作成（company_admin / super_admin）
//	PUT    /api/v2/teaching-materials/:id  更新（同上）
//	DELETE /api/v2/teaching-materials/:id  削除（同上）
type TeachingMaterialHandler struct {
	uc *usecase.TeachingMaterialUseCase
}

func NewTeachingMaterialHandler(uc *usecase.TeachingMaterialUseCase) *TeachingMaterialHandler {
	return &TeachingMaterialHandler{uc: uc}
}

func (h *TeachingMaterialHandler) actorContext(c *gin.Context) (uint64, uint64, string, bool) {
	user := middleware.CurrentUserFromContext(c)
	if user == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return 0, 0, "", false
	}
	companyID := uint64(0)
	if user.CompanyID != nil {
		companyID = *user.CompanyID
	}
	return user.ID, companyID, user.Role, true
}

// List は GET /api/v2/teaching-materials 。
func (h *TeachingMaterialHandler) List(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	rows, err := h.uc.List(c.Request.Context(), companyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "教材の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// Get は GET /api/v2/teaching-materials/:id 。
func (h *TeachingMaterialHandler) Get(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, err := h.uc.Get(c.Request.Context(), id, companyID, role)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "教材が見つかりません"})
			return
		}
		// usecase が "forbidden" を返したら 403。
		if err.Error() == "forbidden" {
			c.JSON(http.StatusForbidden, gin.H{"error": "閲覧権限がありません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "教材の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, m)
}

type teachingMaterialRequest struct {
	Title       string `json:"title"`
	Content     string `json:"content"`
	IsPublished bool   `json:"isPublished"`
}

// Create は POST /api/v2/teaching-materials 。
func (h *TeachingMaterialHandler) Create(c *gin.Context) {
	uid, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	var req teachingMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.uc.Create(c.Request.Context(), usecase.CreateTeachingMaterialInput{
		ActorUserID:    uid,
		ActorCompanyID: companyID,
		ActorRole:      role,
		Title:          req.Title,
		Content:        req.Content,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondTeachingMaterialErr(c, err, "教材の作成に失敗しました")
		return
	}
	c.JSON(http.StatusCreated, m)
}

// Update は PUT /api/v2/teaching-materials/:id 。
func (h *TeachingMaterialHandler) Update(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req teachingMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.uc.Update(c.Request.Context(), usecase.UpdateTeachingMaterialInput{
		ID:             id,
		ActorCompanyID: companyID,
		ActorRole:      role,
		Title:          req.Title,
		Content:        req.Content,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondTeachingMaterialErr(c, err, "教材の更新に失敗しました")
		return
	}
	c.JSON(http.StatusOK, m)
}

// Delete は DELETE /api/v2/teaching-materials/:id 。
func (h *TeachingMaterialHandler) Delete(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, companyID, role); err != nil {
		respondTeachingMaterialErr(c, err, "教材の削除に失敗しました")
		return
	}
	c.Status(http.StatusNoContent)
}

func respondTeachingMaterialErr(c *gin.Context, err error, fallback string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "教材が見つかりません"})
		return
	}
	if err.Error() == "forbidden" || err.Error() == "forbidden: only company_admin or super_admin can create materials" || err.Error() == "actor must belong to a company" {
		c.JSON(http.StatusForbidden, gin.H{"error": "操作権限がありません"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}
