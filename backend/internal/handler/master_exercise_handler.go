package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	repository "github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// MasterExerciseHandler は運営マスタ演習問題の一覧 / 詳細を返す（詳細は slug ベース URL）。
type MasterExerciseHandler struct {
	listExercises  *usecase.ListMasterExercisesUseCase
	listWithStatus *usecase.ListMasterExercisesWithStatusUseCase
	getExercise    *usecase.GetMasterExerciseUseCase
}

func NewMasterExerciseHandler(
	list *usecase.ListMasterExercisesUseCase,
	listWithStatus *usecase.ListMasterExercisesWithStatusUseCase,
	get *usecase.GetMasterExerciseUseCase,
) *MasterExerciseHandler {
	return &MasterExerciseHandler{listExercises: list, listWithStatus: listWithStatus, getExercise: get}
}

// masterExerciseListItemResponse は一覧カード表示に必要な最小フィールドのみを持つレスポンス型。
// description / starterCode / hintText / expectedOutput / explanation などの重いフィールドは
// 詳細 API (GetBySlug) でのみ返す。
type masterExerciseListItemResponse struct {
	ID          uint64 `json:"id"`
	Slug        string `json:"slug"`
	Language    string `json:"language"`
	OrderIndex  int    `json:"orderIndex"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	Difficulty  int16  `json:"difficulty"`
	Mode        string `json:"mode"`
	IsPublished bool   `json:"isPublished"`
	// Status は current user の提出状況。"solved" / "in_progress" / ""（未提出）。
	Status string                            `json:"status"`
	Stats  repository.ExerciseSubmissionStats `json:"stats"`
}

// List は演習問題一覧を query language で絞り込んで返す。
// current user の提出状況と全ユーザ集計を同時に返す（未ログイン時は status 空）。
//
//	@Summary      演習問題 一覧 (status + stats 付き)
//	@Description  運営 マスタ 演習問題 を 取得。 query language で 絞り込み (例: php / go)。 current user の 提出 状況 と 全 user 集計 を 同時 に 返す。 未 ログイン 時 は status 空。
//	@Tags         exercises
//	@Produce      json
//	@Param        language  query     string  false  "言語 フィルタ (例: php, go, javascript)"
//	@Success      200       {array}   masterExerciseListItemResponse
//	@Failure      500       {object}  errorResponse  "DB / 集計 失敗"
//	@Router       /exercises [get]
//	@Security     CookieAuth
func (h *MasterExerciseHandler) List(c *gin.Context) {
	language := c.Query("language")
	uid := middleware.CurrentUserIDOrZero(c)
	rows, err := h.listWithStatus.Execute(c.Request.Context(), usecase.ListMasterExercisesWithStatusInput{
		UserID:   uid,
		Language: language,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}
	items := make([]masterExerciseListItemResponse, len(rows))
	for i, r := range rows {
		items[i] = masterExerciseListItemResponse{
			ID:          r.ID,
			Slug:        r.Slug,
			Language:    r.Language,
			OrderIndex:  r.OrderIndex,
			Category:    r.Category,
			Title:       r.Title,
			Difficulty:  r.Difficulty,
			Mode:        r.Mode,
			IsPublished: r.IsPublished,
			Status:      r.Status,
			Stats:       r.Stats,
		}
	}
	c.JSON(http.StatusOK, items)
}

// GetBySlug は入出力例を含む詳細を返す。
// NotFound だけ 404、それ以外の DB エラーは 500 にして本物の障害を「該当なし」と誤検知させない。
//
//	@Summary      演習問題 詳細 (slug)
//	@Description  指定 slug の 演習 問題 + 入 出力 例 一覧 を 返す。 詳細 画面 用。
//	@Tags         exercises
//	@Produce      json
//	@Param        slug  path      string  true  "問題 slug (例: php-1)"
//	@Success      200   {object}  usecase.GetMasterExerciseDetailOutput
//	@Failure      400   {object}  errorResponse  "slug 欠落"
//	@Failure      404   {object}  errorResponse  "問題 が 見つから ない"
//	@Failure      500   {object}  errorResponse  "DB / 集計 失敗"
//	@Router       /exercises/{slug} [get]
//	@Security     CookieAuth
func (h *MasterExerciseHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	detail, err := h.getExercise.ExecuteBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, detail)
}
