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

// @Summary      コード サンドボックス 実行
// @Description  trainee が 書いた コード を サーバ 側 sandbox で 実行 し stdout/stderr/exitCode を 返す。 language は php / go / bash。
// @Tags         code-execution
// @Accept       json
// @Produce      json
// @Param        body  body      codeExecuteRequest  true  "コード + 言語"
// @Success      200   {object}  domain.CodeExecutionResult
// @Failure      400   {object}  errorResponse  "バリデーション or 実行 失敗"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Router       /code/execute [post]
// @Security     CookieAuth
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

// @Summary      実行環境 ウォームアップ
// @Description  コードエディタ 入場 時 に 呼び、 指定 言語 の 実行 環境 を 事前 に 温める（Go は コンパイル キャッシュ、 php/bash は no-op）。 実行時 に 起動 する のではなく、 入場 時 に warm に する。
// @Tags         code-execution
// @Accept       json
// @Produce      json
// @Param        body  body      codeWarmupRequest  true  "言語"
// @Success      200   {object}  codeWarmupResponse
// @Failure      400   {object}  errorResponse  "バリデーション 失敗"
// @Failure      401   {object}  errorResponse  "未 認証"
// @Router       /code/warmup [post]
// @Security     CookieAuth
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
