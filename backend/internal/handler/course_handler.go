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

// @Summary      コース 一覧
// @Description  current user の role / company で 自動 フィルタ。 trainee は published のみ、 admin 系 は draft 含む。
// @Tags         courses
// @Produce      json
// @Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.Course
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /courses [get]
// @Security     CookieAuth
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

// @Summary      コース 詳細
// @Description  指定 id の コース を 返す。 他社 / 未 公開 (trainee 不可) は 403。
// @Tags         courses
// @Produce      json
// @Param        id  path      int  true  "コース ID"
// @Success      200  {object}  github_com_norman6464_FreStyle_backend_internal_domain.Course
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "操作 権限 なし"
// @Failure      404  {object}  errorResponse  "コース が ない"
// @Router       /courses/{id} [get]
// @Security     CookieAuth
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

// @Summary      コース 作成
// @Description  company_admin / super_admin の み。 CompanyAdmin は 自社 固定。
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        body  body      courseRequest  true  "作成 内容"
// @Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.Course
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      403   {object}  errorResponse  "操作 権限 なし"
// @Router       /courses [post]
// @Security     CookieAuth
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

// @Summary      コース 更新
// @Description  指定 id を 更新 (company_admin / super_admin)。
// @Tags         courses
// @Accept       json
// @Produce      json
// @Param        id    path      int            true  "コース ID"
// @Param        body  body      courseRequest  true  "更新 内容"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.Course
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      403   {object}  errorResponse  "操作 権限 なし"
// @Failure      404   {object}  errorResponse  "コース が ない"
// @Router       /courses/{id} [put]
// @Security     CookieAuth
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

// @Summary      コース 削除
// @Description  指定 id を 削除 + 配下 教材 も cascade 削除 (company_admin / super_admin)。
// @Tags         courses
// @Produce      json
// @Param        id  path  int  true  "コース ID"
// @Success      204  "成功 (本文 なし)"
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "操作 権限 なし"
// @Failure      404  {object}  errorResponse  "コース が ない"
// @Router       /courses/{id} [delete]
// @Security     CookieAuth
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
