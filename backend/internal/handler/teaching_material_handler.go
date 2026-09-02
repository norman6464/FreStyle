package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// TeachingMaterialHandler は教材の CRUD + コース内一覧 API を扱う。
type TeachingMaterialHandler struct {
	uc *usecase.TeachingMaterialUseCase
}

func NewTeachingMaterialHandler(uc *usecase.TeachingMaterialUseCase) *TeachingMaterialHandler {
	return &TeachingMaterialHandler{uc: uc}
}

// ListByCourse はコース配下の教材を返す（path の :id はコース ID）。
//
//	@Summary      コース内 教材 一覧
//	@Description  指定 コース 配下 の 教材 を 返す。 下書き が 混ざる の は その コース を 編集 できる 場合 だけ。 コース を 読め ない 相手 に は 404。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	courseID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course id"})
		return
	}
	rows, err := h.uc.ListByCourse(c.Request.Context(), courseID, usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace})
	if err != nil {
		respondEntityErr(c, err, "教材が見つかりません", "教材の取得に失敗しました")
		return
	}
	c.JSON(http.StatusOK, rows)
}

// @Summary      教材 詳細
// @Description  指定 id の 教材 を 返す。 公開 済み は ワークスペース の 一員 なら 誰 でも 読める。 読め ない 相手 に は、 存在 し ない 場合 と 同じ 404 を 返す。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	m, err := h.uc.Get(c.Request.Context(), id, usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "教材が見つかりません"})
			return
		}
		if errors.Is(err, usecase.ErrMaterialForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "閲覧権限がありません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "教材の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, toChapterDetailResponse(m))
}

// chapterDetailResponse は章の詳細応答。リッチ本文（doc）を JSON のまま返すために
// domain を包む（domain 側は json:"-" で直接は出さない）。未移行の章は doc が null。
type chapterDetailResponse struct {
	*domain.TeachingMaterial
	Doc json.RawMessage `json:"doc"`
}

func toChapterDetailResponse(m *domain.TeachingMaterial) chapterDetailResponse {
	var doc json.RawMessage
	if m.Doc != nil {
		doc = json.RawMessage(*m.Doc)
	}
	return chapterDetailResponse{TeachingMaterial: m, Doc: doc}
}

// updateChapterDocRequest は章のリッチ本文（tiptap JSON）更新の入力。
type updateChapterDocRequest struct {
	Doc              json.RawMessage `json:"doc" binding:"required"`
	ExpectedRevision int             `json:"expectedRevision" binding:"required,min=1"`
}

// UpdateDoc は章のリッチ本文を楽観ロックで保存する。
//
//	@Summary      章 リッチ 本文 更新
//	@Description  tiptap の doc JSON を revision 楽観 ロック で 保存 する。 不一致 は 409。 canManage のみ。
//	@Tags         teaching-materials
//	@Accept       json
//	@Produce      json
//	@Param        id    path      int                     true  "章 ID"
//	@Param        body  body      updateChapterDocRequest true  "doc と expectedRevision"
//	@Success      200   {object}  chapterDetailResponse
//	@Failure      400   {object}  errorResponse  "doc が 不正"
//	@Failure      401   {object}  errorResponse  "未 認証"
//	@Failure      403   {object}  errorResponse  "権限 なし"
//	@Failure      404   {object}  errorResponse  "章 が 見つからない"
//	@Failure      409   {object}  errorResponse  "版 不一致 (他 で 更新 済み)"
//	@Router       /teaching-materials/{id}/doc [put]
//	@Security     CookieAuth
func (h *TeachingMaterialHandler) UpdateDoc(c *gin.Context) {
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateChapterDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.uc.UpdateDoc(c.Request.Context(), usecase.UpdateChapterDocInput{
		MaterialActor:    usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		ID:               id,
		Doc:              string(req.Doc),
		ExpectedRevision: req.ExpectedRevision,
	})
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "教材が見つかりません"})
		case errors.Is(err, repository.ErrChapterDocConflict):
			c.JSON(http.StatusConflict, gin.H{"error": "他の場所で更新されています。最新版を読み込み直してください"})
		case errors.Is(err, usecase.ErrChapterDocInvalid), errors.Is(err, repository.ErrChapterDocInvalidData):
			c.JSON(http.StatusBadRequest, gin.H{"error": "本文の形式が不正です"})
		case errors.Is(err, usecase.ErrMaterialForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "編集権限がありません"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "教材の更新に失敗しました"})
		}
		return
	}
	c.JSON(http.StatusOK, toChapterDetailResponse(m))
}

type teachingMaterialCreateRequest struct {
	CourseID      uint64 `json:"courseId" binding:"required"`
	Title         string `json:"title"`
	OrderInCourse int    `json:"orderInCourse"`
	IsPublished   bool   `json:"isPublished"`
}

type teachingMaterialUpdateRequest struct {
	Title         string `json:"title"`
	OrderInCourse int    `json:"orderInCourse"`
	IsPublished   bool   `json:"isPublished"`
}

// @Summary      教材 作成
// @Description  その コース を 編集 できる 人 だけ。 courseId 必須。 コース を 読め ない 相手 に は 404、 読める が 編集 でき ない 場合 は 403。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
	if !ok {
		return
	}
	var req teachingMaterialCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m, err := h.uc.Create(c.Request.Context(), usecase.CreateTeachingMaterialInput{
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		CourseID:      req.CourseID,
		Title:         req.Title,
		OrderInCourse: req.OrderInCourse,
		IsPublished:   req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "教材が見つかりません", "教材の作成に失敗しました")
		return
	}
	c.JSON(http.StatusCreated, m)
}

// @Summary      教材 更新
// @Description  その 教材 を 編集 できる 人 だけ。 編集 の 可否 は 対象 ごと の 付与 が 決める。 読め ない 相手 に は 404、 読める が 権限 が 足り ない 場合 は 403。
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
	uid, actorWorkspace, _, ok := actorWorkspaceFromContext(c)
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
		ID:            id,
		MaterialActor: usecase.MaterialActor{ActorUserID: uid, ActorWorkspace: actorWorkspace},
		Title:         req.Title,
		OrderInCourse: req.OrderInCourse,
		IsPublished:   req.IsPublished,
	})
	if err != nil {
		respondEntityErr(c, err, "教材が見つかりません", "教材の更新に失敗しました")
		return
	}
	c.JSON(http.StatusOK, m)
}

// @Summary      教材 削除
// @Description  その 教材 を 編集 できる 人 だけ。 読め ない 相手 に は 404、 読める が 権限 が 足り ない 場合 は 403。
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
		respondEntityErr(c, err, "教材が見つかりません", "教材の削除に失敗しました")
		return
	}
	c.Status(http.StatusNoContent)
}
