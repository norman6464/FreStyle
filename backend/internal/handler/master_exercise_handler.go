package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"gorm.io/gorm"
)

// MasterExerciseHandler は運営マスタ演習問題（master_exercises）の一覧 / 取得を返す。
//
// API パスは言語非依存:
//
//	GET /api/v2/exercises?language=php   一覧
//	GET /api/v2/exercises/:slug          詳細（入出力例の配列を含む）
//
// 詳細は slug ベース URL（`/code-editor/php-1` のように人間可読）に揃えるため、
// 旧 `:id` ルートは廃止し slug 必須にする（フロントは新 API に追従させる）。
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

// List は GET /api/v2/exercises 。query `language` で絞り込み（空なら全言語）。
//
// current user の提出状況（solved / in_progress / 未着手）と
// 全ユーザ合計の集計（提出数 / 正答ユーザ数）を 1 度に返す。
// 一覧ページでステータスバッジ + 解答率を出すために必要。
// 未ログインの場合は status は全 ""、 stats だけ返す。
//
//	@Summary      演習問題 一覧 (status + stats 付き)
//	@Description  運営 マスタ 演習問題 を 取得。 query language で 絞り込み (例: php / go)。 current user の 提出 状況 と 全 user 集計 を 同時 に 返す。 未 ログイン 時 は status 空。
//	@Tags         exercises
//	@Produce      json
//	@Param        language  query     string  false  "言語 フィルタ (例: php, go, javascript)"
//	@Success      200       {array}   usecase.MasterExerciseWithStatus
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
	c.JSON(http.StatusOK, rows)
}

// GetBySlug は GET /api/v2/exercises/:slug 。 入出力例を含む詳細を返す。
//
// `gorm.ErrRecordNotFound` のときだけ 404 を返し、 それ以外の DB / 集計エラーは 500 として
// 区別する。 全部 404 にすると本物の障害を「該当なし」と誤検知して上流に隠してしまうため。
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
