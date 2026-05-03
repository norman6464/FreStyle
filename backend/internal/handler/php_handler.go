package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// PhpHandler は PHP 演習問題一覧取得・コード実行の HTTP ハンドラ。
type PhpHandler struct {
	listExercises *usecase.ListPhpExercisesUseCase
	getExercise   *usecase.GetPhpExerciseUseCase
	executeCode   *usecase.ExecuteCodeUseCase
}

func NewPhpHandler(
	list *usecase.ListPhpExercisesUseCase,
	get *usecase.GetPhpExerciseUseCase,
	exec *usecase.ExecuteCodeUseCase,
) *PhpHandler {
	return &PhpHandler{listExercises: list, getExercise: get, executeCode: exec}
}

func (h *PhpHandler) ListExercises(c *gin.Context) {
	exercises, err := h.listExercises.Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, exercises)
}

func (h *PhpHandler) GetExercise(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	exercise, err := h.getExercise.Execute(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, exercise)
}

type executeRequest struct {
	Code     string `json:"code"     binding:"required"`
	Language string `json:"language"`
}

func (h *PhpHandler) Execute(c *gin.Context) {
	var req executeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Language == "" {
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
