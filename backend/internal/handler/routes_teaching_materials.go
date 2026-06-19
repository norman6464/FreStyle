package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// registerTeachingMaterialRoutes は教材個別 API（詳細 / 作成 / 更新 / 削除）を登録する。
// コース配下の教材一覧は registerCourseRoutes 側で登録する。
// アクセス制御は usecase 層で actor の company_id / role を検証する。
func registerTeachingMaterialRoutes(g *gin.RouterGroup, deps *routeDeps) {
	courseRepo := persistence.NewCourseRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)
	chapterViewRepo := persistence.NewUserChapterViewRepository(deps.db)
	uc := usecase.NewTeachingMaterialUseCase(materialRepo, courseRepo)
	h := NewTeachingMaterialHandler(uc)
	g.GET("/teaching-materials", h.List) // backward-compat（frontend のコース対応後に削除予定）
	g.GET("/teaching-materials/:id", h.Get)
	g.POST("/teaching-materials", h.Create)
	g.PUT("/teaching-materials/:id", h.Update)
	g.DELETE("/teaching-materials/:id", h.Delete)

	// 章閲覧記録（「続きから」カードの基盤）。ベストエフォートなので失敗しても 204 を返す。
	cvh := NewChapterViewHandler(usecase.NewRecordChapterViewUseCase(chapterViewRepo, materialRepo, courseRepo))
	g.POST("/teaching-materials/:id/view", cvh.RecordView)
}
