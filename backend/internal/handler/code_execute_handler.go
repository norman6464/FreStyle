package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// CodeExecuteHandler は trainee が書いたコードをサーバ側サンドボックスで実行する。
//
//	POST /api/v2/code/execute
//	Body: { "code": "...", "language": "php" }
//
// 旧 `/api/v2/php/execute` から汎用化した（body に language を持たせる設計は元から
// あったので、path 側を言語非依存に揃えた）。
type CodeExecuteHandler struct {
	executeCode *usecase.ExecuteCodeUseCase
}

func NewCodeExecuteHandler(exec *usecase.ExecuteCodeUseCase) *CodeExecuteHandler {
	return &CodeExecuteHandler{executeCode: exec}
}

type codeExecuteRequest struct {
	Code     string `json:"code"     binding:"required"`
	Language string `json:"language"`
}

func (h *CodeExecuteHandler) Execute(c *gin.Context) {
	var req codeExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Language == "" {
		// language 未指定はデフォルト php（既存挙動維持）。多言語対応で trainee が
		// 明示的に言語を切り替える UI ができるまでの暫定措置。
		req.Language = "php"
	}

	out, err := h.executeCode.Execute(c.Request.Context(), usecase.ExecuteCodeInput{
		Code:     req.Code,
		Language: req.Language,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}
