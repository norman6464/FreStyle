package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// MasterExerciseHandler は運営マスタ演習問題（master_exercises）の一覧 / 取得を返す。
//
// API パスは言語非依存:
//
//	GET /api/v2/exercises?language=php
//	GET /api/v2/exercises/:id
//
// 旧 `/api/v2/php/exercises` 系は撤去（フロントは新 API に追従させる方針）。
type MasterExerciseHandler struct {
	listExercises *usecase.ListMasterExercisesUseCase
	getExercise   *usecase.GetMasterExerciseUseCase
}

func NewMasterExerciseHandler(
	list *usecase.ListMasterExercisesUseCase,
	get *usecase.GetMasterExerciseUseCase,
) *MasterExerciseHandler {
	return &MasterExerciseHandler{listExercises: list, getExercise: get}
}

// List は GET /api/v2/exercises 。query `language` で絞り込み（空なら全言語）。
func (h *MasterExerciseHandler) List(c *gin.Context) {
	language := c.Query("language")
	exercises, err := h.listExercises.Execute(language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "演習問題の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, exercises)
}

// Get は GET /api/v2/exercises/:id 。
func (h *MasterExerciseHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	exercise, err := h.getExercise.Execute(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "演習問題が見つかりません"})
		return
	}
	c.JSON(http.StatusOK, exercise)
}
