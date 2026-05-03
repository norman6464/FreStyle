package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerPhpRoutes は PHP 学習環境の REST エンドポイントを登録する。
func registerPhpRoutes(g *gin.RouterGroup, deps *routeDeps) {
	exerciseRepo := repository.NewPhpExerciseRepository(deps.db)
	h := NewPhpHandler(
		usecase.NewListPhpExercisesUseCase(exerciseRepo),
		usecase.NewGetPhpExerciseUseCase(exerciseRepo),
		usecase.NewExecuteCodeUseCase(),
	)
	g.GET("/php/exercises", h.ListExercises)
	g.GET("/php/exercises/:id", h.GetExercise)
	g.POST("/php/execute", h.Execute)
}
