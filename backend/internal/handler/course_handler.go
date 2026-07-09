package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// CourseHandler はコースの CRUD + 最終閲覧章 API を扱う。
type CourseHandler struct {
	uc               *usecase.CourseUseCase
	listWithProgress *usecase.ListCoursesWithProgressUseCase
	lastViewed       *usecase.GetLastViewedChapterUseCase
}

func NewCourseHandler(
	uc *usecase.CourseUseCase,
	listWithProgress *usecase.ListCoursesWithProgressUseCase,
	lastViewed *usecase.GetLastViewedChapterUseCase,
) *CourseHandler {
	return &CourseHandler{uc: uc, listWithProgress: listWithProgress, lastViewed: lastViewed}
}

// @Summary      コース 一覧 (進捗付き)
// @Description  current user の role / company で 自動 フィルタ。 trainee は published のみ、 admin 系 は draft 含む。 各コース に 章数 materialCount と 自身 の 完了 章数 completedCount を 付与 して 返す。
// @Tags         courses
// @Produce      json
// @Success      200  {array}   usecase.CourseWithProgress
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /courses [get]
// @Security     CookieAuth
func (h *CourseHandler) List(c *gin.Context) {
	uid, companyID, role, ok := actorFromContext(c)
	if !ok {
		return
	}
	rows, err := h.listWithProgress.Execute(c.Request.Context(), usecase.ListCoursesWithProgressInput{
		ActorUserID:    uid,
		ActorCompanyID: companyID,
		ActorRole:      role,
	})
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
	_, companyID, role, ok := actorFromContext(c)
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
		respondEntityErr(c, err, "コースが見つかりません", "コースの取得に失敗しました")
		return
	}
	c.JSON(http.StatusOK, course)
}

// LastViewed は current user がコース内で最後に閲覧した章の閲覧記録を返す。
//
// @Summary      コース内の最終閲覧章
// @Description  current user が この コース で 最後 に 閲覧 した 章 の 閲覧 記録 を 返す。 コース詳細 の 「続き から 表示」 用。 履歴 なし は 204。
// @Tags         courses
// @Produce      json
// @Param        id  path      int  true  "コース ID"
// @Success      200  {object}  github_com_norman6464_FreStyle_backend_internal_domain.UserChapterView
// @Success      204  "履歴 なし (本文 なし)"
// @Failure      400  {object}  errorResponse  "id 不正"
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      403  {object}  errorResponse  "閲覧 権限 なし"
// @Failure      404  {object}  errorResponse  "コース が ない"
// @Router       /courses/{id}/last-viewed [get]
// @Security     CookieAuth
func (h *CourseHandler) LastViewed(c *gin.Context) {
	uid, companyID, role, ok := actorFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	view, err := h.lastViewed.Execute(c.Request.Context(), usecase.GetLastViewedChapterInput{
		UserID:         uid,
		ActorCompanyID: companyID,
		ActorRole:      role,
		CourseID:       id,
	})
	if err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "閲覧履歴の取得に失敗しました")
		return
	}
	if view == nil {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, view)
}

type courseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// Category は定義済みの学習領域のみ許可(空 = 未分類)。値の正本は domain.ValidCourseCategories。
	// oneof で宣言的に 400 を返し、usecase 側でも防衛的に検証する。
	Category    string `json:"category" binding:"omitempty,oneof=dev-basics backend architecture database infra security product"`
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
	uid, companyID, role, ok := actorFromContext(c)
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
		Category:       req.Category,
		SortOrder:      req.SortOrder,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの作成に失敗しました")
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
	_, companyID, role, ok := actorFromContext(c)
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
		Category:       req.Category,
		SortOrder:      req.SortOrder,
		IsPublished:    req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの更新に失敗しました")
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
	_, companyID, role, ok := actorFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, companyID, role); err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの削除に失敗しました")
		return
	}
	c.Status(http.StatusNoContent)
}
