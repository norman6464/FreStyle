package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerCourseRoutes はコース CRUD + コース配下教材一覧 API を登録する。
// アクセス制御は usecase 層で actor の company_id / role を検証する。
func registerCourseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	courseRepo := persistence.NewCourseRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)

	courseUC := usecase.NewCourseUseCase(courseRepo, materialRepo)
	listWithCountUC := usecase.NewListCoursesWithMaterialCountUseCase(courseRepo, materialRepo)
	courseHandler := NewCourseHandler(courseUC, listWithCountUC)

	materialUC := usecase.NewTeachingMaterialUseCase(materialRepo, courseRepo)
	materialHandler := NewTeachingMaterialHandler(materialUC)

	g.GET("/courses", courseHandler.List)
	g.GET("/courses/:id", courseHandler.Get)
	g.POST("/courses", courseHandler.Create)
	g.PUT("/courses/:id", courseHandler.Update)
	g.DELETE("/courses/:id", courseHandler.Delete)
	g.GET("/courses/:id/materials", materialHandler.ListByCourse)
}
