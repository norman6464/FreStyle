package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// CodeExecuteHandler は trainee が書いたコードをサーバ側サンドボックスで実行する。
// エディタ入場時の事前ウォームアップ（warmup）も担う。
type CodeExecuteHandler struct {
	executeCode *usecase.ExecuteCodeUseCase
	warmupCode  *usecase.WarmupCodeUseCase
}

// NewCodeExecuteHandler は実行 / ウォームアップ usecase を注入して handler を返す。
func NewCodeExecuteHandler(exec *usecase.ExecuteCodeUseCase, warmup *usecase.WarmupCodeUseCase) *CodeExecuteHandler {
	return &CodeExecuteHandler{executeCode: exec, warmupCode: warmup}
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
		// language 未指定はデフォルト php（既存挙動維持）。
		req.Language = "php"
	}

	out, err := h.executeCode.Execute(c.Request.Context(), domain.CodeExecutionInput{
		Code:     req.Code,
		Language: req.Language,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

type codeWarmupRequest struct {
	Language string `json:"language" binding:"required"`
}

// codeWarmupResponse はウォームアップ結果。ready=true で実行環境が準備済み。
type codeWarmupResponse struct {
	Ready bool `json:"ready"`
}

func (h *CodeExecuteHandler) Warmup(c *gin.Context) {
	var req codeWarmupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.warmupCode.Execute(c.Request.Context(), req.Language); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, codeWarmupResponse{Ready: true})
}
