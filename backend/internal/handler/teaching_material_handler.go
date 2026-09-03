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
