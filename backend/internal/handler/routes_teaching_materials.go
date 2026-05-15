package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerTeachingMaterialRoutes は教材個別 API を登録する:
//
//	GET    /api/v2/teaching-materials/:id  詳細
//	POST   /api/v2/teaching-materials      作成（company_admin / super_admin、 body の courseId 必須）
//	PUT    /api/v2/teaching-materials/:id  更新（同上）
//	DELETE /api/v2/teaching-materials/:id  削除（同上）
//
// コース配下の教材一覧 (`GET /api/v2/courses/:id/materials`) は
// registerCourseRoutes 側で登録する（コース ACL の参照を簡潔にするため）。
//
// アクセス制御は usecase 層で actor の company_id / role を検証する。
// trainee は同一 company の `is_published=true` 教材かつ所属コースが published のときのみ閲覧可。
func registerTeachingMaterialRoutes(g *gin.RouterGroup, deps *routeDeps) {
	courseRepo := persistence.NewCourseRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)
	uc := usecase.NewTeachingMaterialUseCase(materialRepo, courseRepo)
	h := NewTeachingMaterialHandler(uc)
	g.GET("/teaching-materials", h.List) // backward-compat（frontend のコース対応後に削除予定）
	g.GET("/teaching-materials/:id", h.Get)
	g.POST("/teaching-materials", h.Create)
	g.PUT("/teaching-materials/:id", h.Update)
	g.DELETE("/teaching-materials/:id", h.Delete)
}
