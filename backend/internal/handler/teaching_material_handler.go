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
//	GET    /api/v2/courses/:courseId/materials      コース内一覧（actor role で自動フィルタ）
//	GET    /api/v2/teaching-materials/:id           詳細
//	POST   /api/v2/teaching-materials               作成（company_admin / super_admin、 course_id 必須）
//	PUT    /api/v2/teaching-materials/:id           更新（同上）
//	DELETE /api/v2/teaching-materials/:id           削除（同上）
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

// List は backward-compat 用 GET /api/v2/teaching-materials 。 frontend がコース対応に
// 切り替わったら削除予定。
//
//	@Summary      教材 全 件 一覧 (deprecated)
//	@Description  backward-compat 用。 company 内 全 教材 を 返す。 frontend が コース 対応 完了 後 に 削除 予定。
//	@Tags         teaching-materials
//	@Produce      json
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.TeachingMaterial
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /teaching-materials [get]
//	@Security     CookieAuth
//	@Deprecated
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

// ListByCourse は GET /api/v2/courses/:id/materials 。 コース配下の教材を返す。
// path param `:id` はコース ID（ルート競合を避けるため materials path 側で `:id` を流用）。
//
//	@Summary      コース内 教材 一覧
//	@Description  指定 コース 配下 の 教材 を 返す。 trainee は published のみ。
//	@Tags         teaching-materials
//	@Produce      json
//	@Param        id  path      int  true  "コース ID"
//	@Success      200  {array}   github_com_norman6464_FreStyle_backend_internal_domain.TeachingMaterial
//	@Failure      400  {object}  errorResponse  "course id 不正"
//	@Failure      401  {object}  errorResponse  "未 認証"
//	@Failure      403  {object}  errorResponse  "他社 コース"
//	@Failure      500  {object}  errorResponse  "DB 失敗"
//	@Router       /courses/{id}/materials [get]
//	@Security     CookieAuth
func (h *TeachingMaterialHandler) ListByCourse(c *gin.Context) {
	_, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}
	rows, err := h.uc.ListByCourse(c.Request.Context(), courseID, companyID, role)
	if err != nil {
		respondTeachingMaterialErr(c, err, "教材の取得に失敗しました")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// @Summary      教材 詳細
// @Description  指定 id の 教材 を 返す。 他社 / 未 公開 (trainee) は 403。
// @Tags         teaching-materials
// @Produce      json
// @Param        id  path      int  true  "教材 ID"
// @Success      200  {object}  github_com_norman6464_FreStyle_backend_internal_domain.TeachingMaterial
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "操作 権限 なし"
// @Failure      404  {object}  errorResponse  "教材 が ない"
// @Router       /teaching-materials/{id} [get]
// @Security     CookieAuth
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

type teachingMaterialCreateRequest struct {
	CourseID      uint64 `json:"courseId" binding:"required"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	OrderInCourse int    `json:"orderInCourse"`
	IsPublished   bool   `json:"isPublished"`
}

type teachingMaterialUpdateRequest struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	OrderInCourse int    `json:"orderInCourse"`
	IsPublished   bool   `json:"isPublished"`
}

// @Summary      教材 作成
// @Description  company_admin / super_admin の み。 courseId 必須。
// @Tags         teaching-materials
// @Accept       json
// @Produce      json
// @Param        body  body      teachingMaterialCreateRequest  true  "作成 内容"
// @Success      201   {object}  github_com_norman6464_FreStyle_backend_internal_domain.TeachingMaterial
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      403   {object}  errorResponse  "操作 権限 なし"
// @Router       /teaching-materials [post]
// @Security     CookieAuth
func (h *TeachingMaterialHandler) Create(c *gin.Context) {
	uid, companyID, role, ok := h.actorContext(c)
	if !ok {
		return
	}
	var req teachingMaterialCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.uc.Create(c.Request.Context(), usecase.CreateTeachingMaterialInput{
		ActorUserID:    uid,
		ActorCompanyID: companyID,
		ActorRole:      role,
		CourseID:       req.CourseID,
		Title:          req.Title,
		Content:        req.Content,
		OrderInCourse:  req.OrderInCourse,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondTeachingMaterialErr(c, err, "教材の作成に失敗しました")
		return
	}
	c.JSON(http.StatusCreated, m)
}

// @Summary      教材 更新
// @Description  指定 id の 教材 を 更新 (company_admin / super_admin)。
// @Tags         teaching-materials
// @Accept       json
// @Produce      json
// @Param        id    path      int                            true  "教材 ID"
// @Param        body  body      teachingMaterialUpdateRequest  true  "更新 内容"
// @Success      200   {object}  github_com_norman6464_FreStyle_backend_internal_domain.TeachingMaterial
// @Failure      400   {object}  errorResponse  "バリデーション"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Failure      403   {object}  errorResponse  "操作 権限 なし"
// @Failure      404   {object}  errorResponse  "教材 が ない"
// @Router       /teaching-materials/{id} [put]
// @Security     CookieAuth
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
	var req teachingMaterialUpdateRequest
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
		OrderInCourse:  req.OrderInCourse,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondTeachingMaterialErr(c, err, "教材の更新に失敗しました")
		return
	}
	c.JSON(http.StatusOK, m)
}

// @Summary      教材 削除
// @Description  指定 id の 教材 を 削除 (company_admin / super_admin)。
// @Tags         teaching-materials
// @Produce      json
// @Param        id  path  int  true  "教材 ID"
// @Success      204  "成功 (本文 なし)"
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "操作 権限 なし"
// @Failure      404  {object}  errorResponse  "教材 が ない"
// @Router       /teaching-materials/{id} [delete]
// @Security     CookieAuth
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
