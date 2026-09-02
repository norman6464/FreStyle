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
// @Description  見せ て よい コース だけ を 返す。 公開 済み は ワークスペース の 一員 なら 誰 でも、 下書き は その コース を 編集 できる 人 だけ に 見える。 各コース に 章数 materialCount と 自身 の 完了 章数 completedCount を 付与 する (下書き の 章 を 数 に 含める の は、 その コース を 編集 できる 場合 だけ)。
// @Tags         courses
// @Produce      json
// @Success      200  {array}   usecase.CourseWithProgress
// @Failure      401  {object}  errorResponse  "未 認証"
// @Failure      500  {object}  errorResponse  "DB 失敗"
// @Router       /courses [get]
// @Security     CookieAuth
func (h *CourseHandler) List(c *gin.Context) {
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	rows, err := h.listWithProgress.Execute(c.Request.Context(), usecase.ListCoursesWithProgressInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "コースの取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, rows)
}

// @Summary      コース 詳細
// @Description  指定 id の コース を 返す。 公開 済み は ワークスペース の 一員 なら 誰 でも 読める。 読め ない 相手 に は、 存在 し ない 場合 と 同じ 404 を 返す (応答 の 差 から 実在 を 読ま せ ない)。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	course, err := h.uc.Get(c.Request.Context(), id, usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace})
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	view, err := h.lastViewed.Execute(c.Request.Context(), usecase.GetLastViewedChapterInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		CourseID:      id,
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
	Category string `json:"category" binding:"omitempty,oneof=dev-basics backend architecture database infra security product"`
	// Language は主に扱う言語・技術(例: "go" / "docker"。空 = 言語が主題でない)。
	// 演習の language と同じ自由文字列方式(表示色は frontend のカラーマップが持つ)。
	Language    string `json:"language" binding:"omitempty,max=50"`
	SortOrder   int    `json:"sortOrder"`
	IsPublished bool   `json:"isPublished"`
}

// @Summary      コース 作成
// @Description  ワークスペース の 一員 なら 誰 でも 作れる。 作っ た 人 は その コース の admin に なる (コース と 付与 は 同じ トランザクション で 書く)。 未 所属 は 403。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	var req courseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	course, err := h.uc.Create(c.Request.Context(), usecase.CreateCourseInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Language:      req.Language,
		SortOrder:     req.SortOrder,
		IsPublished:   req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの作成に失敗しました")
		return
	}
	c.JSON(http.StatusCreated, course)
}

// @Summary      コース 更新
// @Description  その コース を 編集 できる 人 だけ。 編集 の 可否 は 対象 ごと の 付与 が 決める (アプリ の ロール は 見 ない)。 読め ない 相手 に は 404、 読める が 権限 が 足り ない 場合 は 403。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
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
		ID:            id,
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Language:      req.Language,
		SortOrder:     req.SortOrder,
		IsPublished:   req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの更新に失敗しました")
		return
	}
	c.JSON(http.StatusOK, course)
}

// @Summary      コース 削除
// @Description  その コース を 編集 できる 人 だけ。 配下 の 教材 も 一緒 に 消える。 読め ない 相手 に は 404、 読める が 権限 が 足り ない 場合 は 403。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.Delete(c.Request.Context(), id, usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace}); err != nil {
		respondEntityErr(c, err, "コースが見つかりません", "コースの削除に失敗しました")
		return
	}
	c.Status(http.StatusNoContent)
}
