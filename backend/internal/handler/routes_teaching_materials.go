package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/repository"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerTeachingMaterialRoutes は教材 API 5 本を登録する:
//
//	GET    /api/v2/teaching-materials      一覧
//	GET    /api/v2/teaching-materials/:id  詳細
//	POST   /api/v2/teaching-materials      作成（company_admin / super_admin）
//	PUT    /api/v2/teaching-materials/:id  更新（同上）
//	DELETE /api/v2/teaching-materials/:id  削除（同上）
//
// アクセス制御は usecase 層で actor の company_id / role を検証する。
// trainee は同一 company の `is_published=true` 教材のみ閲覧可。
func registerTeachingMaterialRoutes(g *gin.RouterGroup, deps *routeDeps) {
	repo := repository.NewTeachingMaterialRepository(deps.db)
	uc := usecase.NewTeachingMaterialUseCase(repo)
	h := NewTeachingMaterialHandler(uc)
	g.GET("/teaching-materials", h.List)
	g.GET("/teaching-materials/:id", h.Get)
	g.POST("/teaching-materials", h.Create)
	g.PUT("/teaching-materials/:id", h.Update)
	g.DELETE("/teaching-materials/:id", h.Delete)
}
