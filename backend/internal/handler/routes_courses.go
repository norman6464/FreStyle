package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerCourseRoutes はコース API + コース配下教材一覧 API を登録する。
//
//	GET    /api/v2/courses                          一覧
//	GET    /api/v2/courses/:id                      詳細
//	POST   /api/v2/courses                          作成（company_admin / super_admin）
//	PUT    /api/v2/courses/:id                      更新（同上）
//	DELETE /api/v2/courses/:id                      削除（配下教材も cascade で削除）
//	GET    /api/v2/courses/:courseId/materials      コース内教材一覧
//
// アクセス制御は usecase 層で actor の company_id / role を検証する。
// trainee は同一 company の `is_published=true` コースのみ閲覧可。
func registerCourseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	courseRepo := persistence.NewCourseRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)

	courseUC := usecase.NewCourseUseCase(courseRepo, materialRepo)
	courseHandler := NewCourseHandler(courseUC)

	materialUC := usecase.NewTeachingMaterialUseCase(materialRepo, courseRepo)
	materialHandler := NewTeachingMaterialHandler(materialUC)

	g.GET("/courses", courseHandler.List)
	g.GET("/courses/:id", courseHandler.Get)
	g.POST("/courses", courseHandler.Create)
	g.PUT("/courses/:id", courseHandler.Update)
	g.DELETE("/courses/:id", courseHandler.Delete)
	g.GET("/courses/:id/materials", materialHandler.ListByCourse)
}
