package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	repository "github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// MasterExerciseHandler は運営マスタ演習問題の一覧 / 詳細を返す（詳細は slug ベース URL）。
type MasterExerciseHandler struct {
	listExercises   *usecase.ListMasterExercisesUseCase
	listWithStatus  *usecase.ListMasterExercisesWithStatusUseCase
	getExercise     *usecase.GetMasterExerciseUseCase
	languageSummary *usecase.GetExerciseLanguageSummaryUseCase
}

func NewMasterExerciseHandler(
	list *usecase.ListMasterExercisesUseCase,
	listWithStatus *usecase.ListMasterExercisesWithStatusUseCase,
	get *usecase.GetMasterExerciseUseCase,
	languageSummary *usecase.GetExerciseLanguageSummaryUseCase,
) *MasterExerciseHandler {
	return &MasterExerciseHandler{
		listExercises:   list,
		listWithStatus:  listWithStatus,
		getExercise:     get,
		languageSummary: languageSummary,
	}
}

// exerciseLanguageSummaryResponse は言語ごとの問題数 / 正解済み件数。
type exerciseLanguageSummaryResponse struct {
	Language string `json:"language"`
	Total    int64  `json:"total"`
	Solved   int64  `json:"solved"`
}

// Summary は公開済み演習を言語ごとに集計して返す（コード学習の言語選択カード用）。
func (h *MasterExerciseHandler) Summary(c *gin.Context) {
	uid := middleware.CurrentUserIDOrZero(c)

	rows, err := h.languageSummary.Execute(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習の集計に失敗しました"})
		return
	}

	out := make([]exerciseLanguageSummaryResponse, len(rows))
	for i, r := range rows {
		out[i] = exerciseLanguageSummaryResponse{Language: r.Language, Total: r.Total, Solved: r.Solved}
	}
	c.JSON(http.StatusOK, out)
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
	Status string                             `json:"status"`
	Stats  repository.ExerciseSubmissionStats `json:"stats"`
}

// exercisePageResponse はスクロール型ページネーションのレスポンス。
// Items は limit 件以下の問題リスト。HasNext が true のとき次のページが存在する。
type exercisePageResponse struct {
	Items   []masterExerciseListItemResponse `json:"items"`
	HasNext bool                             `json:"hasNext"`
	Offset  int                              `json:"offset"`
	Limit   int                              `json:"limit"`
}

const (
	exerciseDefaultLimit = 20
	exerciseMaxLimit     = 100
)

// List は演習問題一覧をスクロール型ページネーションで返す。
// limit（デフォルト 20・最大 100）と offset（デフォルト 0）で取得範囲を指定する。
// hasNext が true のとき次ページが存在する。
func (h *MasterExerciseHandler) List(c *gin.Context) {
	language := c.Query("language")
	uid := middleware.CurrentUserIDOrZero(c)

	limit := exerciseDefaultLimit
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = min(n, exerciseMaxLimit)
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	// limit+1 件取得して hasNext を判定する（COUNT クエリを別途発行しない）。
	rows, err := h.listWithStatus.Execute(c.Request.Context(), usecase.ListMasterExercisesWithStatusInput{
		UserID:   uid,
		Language: language,
		Offset:   offset,
		Limit:    limit + 1,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}

	hasNext := len(rows) > limit
	if hasNext {
		rows = rows[:limit]
	}

	items := make([]masterExerciseListItemResponse, len(rows))
	for i, r := range rows {
		items[i] = masterExerciseListItemResponse{
			ID:          r.ID,
			Slug:        r.Slug,
			Language:    r.Language,
			OrderIndex:  r.SortOrder,
			Category:    r.Category,
			Title:       r.Title,
			Difficulty:  r.Difficulty,
			Mode:        r.Mode,
			IsPublished: r.IsPublished,
			Status:      r.Status,
			Stats:       r.Stats,
		}
	}
	c.JSON(http.StatusOK, exercisePageResponse{
		Items:   items,
		HasNext: hasNext,
		Offset:  offset,
		Limit:   limit,
	})
}

// min は Go 1.21+ 組み込みと競合しないよう、ローカル定義は不要。
// Go 1.21+ では builtin の min を使える。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetBySlug は入出力例を含む詳細を返す。
// NotFound だけ 404、それ以外の DB エラーは 500 にして本物の障害を「該当なし」と誤検知させない。
func (h *MasterExerciseHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}
	detail, err := h.getExercise.ExecuteBySlug(c.Request.Context(), slug)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, detail)
}
