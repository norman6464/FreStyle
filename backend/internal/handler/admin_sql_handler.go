package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/handler/middleware"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// AdminSQLHandler は super_admin 専用の read-only SQL コンソールを扱う。
type AdminSQLHandler struct {
	execute *usecase.ExecuteReadOnlySQLUseCase
}

// NewAdminSQLHandler は AdminSQLHandler を生成する。
func NewAdminSQLHandler(execute *usecase.ExecuteReadOnlySQLUseCase) *AdminSQLHandler {
	return &AdminSQLHandler{execute: execute}
}

// isSuperAdmin は actor が super_admin（運営管理者）かを判定する。
func isSuperAdmin(actor *domain.User) bool {
	return actor != nil && actor.Role == domain.RoleSuperAdmin
}

// runSQLRequest は read-only SQL 実行の入力。
type runSQLRequest struct {
	Query string `json:"query" binding:"required"`
}

// runSQLResponse は read-only SQL の実行結果。
type runSQLResponse struct {
	Columns   []string `json:"columns"`
	Rows      [][]any  `json:"rows"`
	RowCount  int      `json:"rowCount"`
	Truncated bool     `json:"truncated"`
}

// Run は super_admin が入力した read-only SQL を実行して結果を返すハンドラ。
//
//	@Summary      read-only SQL を実行（super_admin 専用）
//	@Description  super_admin が入力した単一の SELECT / WITH クエリを read-only トランザクション内で実行し、列・行を返す。書き込み系は DB レベルで拒否され、行数は上限で打ち切られる（truncated=true）。
//	@Tags         admin
//	@Accept       json
//	@Produce      json
//	@Param        body  body      runSQLRequest  true  "実行する SELECT / WITH クエリ"
//	@Success      200   {object}  runSQLResponse
//	@Failure      400   {object}  errorResponse  "クエリが空 / 読み取り専用でない / SQL エラー"
//	@Failure      403   {object}  errorResponse  "super_admin 以外"
//	@Router       /admin/sql [post]
//	@Security     CookieAuth
func (h *AdminSQLHandler) Run(c *gin.Context) {
	actor := middleware.CurrentUserFromContext(c)
	if !isSuperAdmin(actor) {
		c.JSON(http.StatusForbidden, errorResponse{Error: "forbidden"})
		return
	}

	var body runSQLRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: "invalid_request"})
		return
	}

	// 監査: 誰がどの SQL を実行したかを必ず残す。
	log.Printf("AUDIT sql-console: super_admin(id=%d email=%s) query=%q", actor.ID, actor.Email, body.Query)

	result, err := h.execute.Execute(c.Request.Context(), usecase.ExecuteReadOnlySQLInput{Query: body.Query})
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, runSQLResponse{
		Columns:   result.Columns,
		Rows:      result.Rows,
		RowCount:  len(result.Rows),
		Truncated: result.Truncated,
	})
}
