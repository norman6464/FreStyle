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

// CourseHandler はコース API を扱う。
//
//	GET    /api/v2/courses        一覧（actor role / company で自動フィルタ）
//	GET    /api/v2/courses/:id    詳細
//	POST   /api/v2/courses        作成（company_admin / super_admin）
//	PUT    /api/v2/courses/:id    更新（同上）
//	DELETE /api/v2/courses/:id    削除（配下教材も cascade で削除）
type CourseHandler struct {
	uc *usecase.CourseUseCase
}

func NewCourseHandler(uc *usecase.CourseUseCase) *CourseHandler {
	return &CourseHandler{uc: uc}
}

func (h *CourseHandler) actorContext(c *gin.Context) (uint64, uint64, string, bool) {
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

func (h *CourseHandler) List(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	rows, err := h.uc.List(c.Request.Context(), companyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "コースの取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *CourseHandler) Get(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	course, err := h.uc.Get(c.Request.Context(), id, companyID, role)
	if err != nil {
		respondCourseErr(c, err, "コースの取得に失敗しました")
		return
	}
	c.JSON(http.StatusOK, course)
}

type courseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	SortOrder   int    `json:"sortOrder"`
	IsPublished bool   `json:"isPublished"`
}

func (h *CourseHandler) Create(c *gin.Context) {
	uid, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	course, err := h.uc.Create(c.Request.Context(), usecase.CreateCourseInput{
		ActorUserID:    uid,
		ActorCompanyID: companyID,
		ActorRole:      role,
		Title:          req.Title,
		Description:    req.Description,
		SortOrder:      req.SortOrder,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondCourseErr(c, err, "コースの作成に失敗しました")
		return
	}
	c.JSON(http.StatusCreated, course)
}

func (h *CourseHandler) Update(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	course, err := h.uc.Update(c.Request.Context(), usecase.UpdateCourseInput{
		ID:             id,
		ActorCompanyID: companyID,
		ActorRole:      role,
		Title:          req.Title,
		Description:    req.Description,
		SortOrder:      req.SortOrder,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondCourseErr(c, err, "コースの更新に失敗しました")
		return
	}
	c.JSON(http.StatusOK, course)
}

func (h *CourseHandler) Delete(c *gin.Context) {
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
		respondCourseErr(c, err, "コースの削除に失敗しました")
		return
	}
	c.Status(http.StatusNoContent)
}

func respondCourseErr(c *gin.Context, err error, fallback string) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "コースが見つかりません"})
		return
	}
	if err.Error() == "forbidden" || err.Error() == "forbidden: only company_admin or super_admin can create courses" || err.Error() == "actor must belong to a company" {
		c.JSON(http.StatusForbidden, gin.H{"error": "操作権限がありません"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": fallback})
}
