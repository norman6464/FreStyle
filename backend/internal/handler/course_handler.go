package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
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

func (h *CourseHandler) List(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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

// courseDetailResponse はコース詳細の返却形。
//
// usecase の型をそのまま返さないのは、内部の型を変えたときに HTTP の契約が
// 黙って変わらないようにするため。載せる項目はここで明示する。
type courseDetailResponse struct {
	domain.Course
	// CanEdit は書き換えられるか（編集 UI を出すかの判定に使う）。
	//
	// 画面がアプリのロールで判断しないよう、可否はサーバーが答える
	// （ロールで出すと「ボタンは出るのに保存が弾かれる」状態になる）。
	CanEdit bool `json:"canEdit"`
	// CanManage は権限そのものを変えられるか（共有ボタンを出すかの判定に使う）。
	CanManage bool `json:"canManage"`
}

func (h *CourseHandler) Get(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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
	c.JSON(http.StatusOK, courseDetailResponse{
		Course:    course.Course,
		CanEdit:   course.CanEdit,
		CanManage: course.CanManage,
	})
}

// LastViewed は current user がコース内で最後に閲覧した章の閲覧記録を返す。
func (h *CourseHandler) LastViewed(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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

func (h *CourseHandler) Create(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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

func (h *CourseHandler) Update(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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

func (h *CourseHandler) Delete(c *gin.Context) {
	uid, actorWorkspace, ok := actorWorkspaceFromContext(c)
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
