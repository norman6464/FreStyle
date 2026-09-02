package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// registerCourseRoutes はコース CRUD + コース配下教材一覧 API を登録する。
// アクセス制御は usecase 層で actor の workspace_id / role を検証する。
func registerCourseRoutes(g *gin.RouterGroup, deps *routeDeps) {
	courseRepo := persistence.NewCourseRepository(deps.db)
	materialRepo := persistence.NewTeachingMaterialRepository(deps.db)
	progressRepo := persistence.NewLessonProgressRepository(deps.db)
	chapterViewRepo := persistence.NewUserChapterViewRepository(deps.db)

	// 教材の可否は対象ごとの付与だけで決まる。判定はこの 1 つを全経路が共有する
	// （経路ごとに別の判定を持つと、同じ教材の可否が場所によって食い違う）。
	permUC := usecase.NewCheckMaterialPermissionUseCase(persistence.NewMaterialPermissionRepository(deps.db))
	principalRepo := persistence.NewKnowledgeBasePermissionRepository(deps.db)

	courseUC := usecase.NewCourseUseCase(courseRepo, materialRepo, permUC, principalRepo)
	listWithProgressUC := usecase.NewListCoursesWithProgressUseCase(
		materialRepo, progressRepo, persistence.NewMaterialPermissionRepository(deps.db),
	)
	lastViewedUC := usecase.NewGetLastViewedChapterUseCase(chapterViewRepo, permUC)
	courseHandler := NewCourseHandler(courseUC, listWithProgressUC, lastViewedUC)

	materialUC := usecase.NewTeachingMaterialUseCase(materialRepo, courseRepo, permUC)
	materialHandler := NewTeachingMaterialHandler(materialUC)

	g.GET("/courses", courseHandler.List)
	g.GET("/courses/:id", courseHandler.Get)
	g.GET("/courses/:id/last-viewed", courseHandler.LastViewed)
	g.POST("/courses", courseHandler.Create)
	g.PUT("/courses/:id", courseHandler.Update)
	g.DELETE("/courses/:id", courseHandler.Delete)
	g.GET("/courses/:id/materials", materialHandler.ListByCourse)

	// 権限そのものを変える経路。呼べるのはその対象を管理できる人だけで、
	// 読めない相手には（存在しない場合と同じ）404 を返す。
	gh := newMaterialGrantHandler(deps, permUC, principalRepo)
	g.GET("/courses/:id/grants", gh.ListCourseGrants)
	g.PUT("/courses/:id/grants/:principalId", gh.GrantCourseRole)
	g.DELETE("/courses/:id/grants/:principalId", gh.RevokeCourseRole)
	// 相手選び。中身はワークスペース全体だが、呼べるかはコース単位で決まる。
	g.GET("/courses/:id/principals", gh.ListGrantablePrincipals)
}

// newMaterialGrantHandler は教材の権限操作 handler を組み立てる。
// コース側と教材側の両方のルートが同じ 1 つを使う（判定を 2 つ持たない）。
func newMaterialGrantHandler(
	deps *routeDeps,
	permUC *usecase.CheckMaterialPermissionUseCase,
	principals repository.KnowledgeBasePermissionRepository,
) *MaterialGrantHandler {
	perm := persistence.NewMaterialPermissionRepository(deps.db)
	return NewMaterialGrantHandler(
		usecase.NewGrantCourseRoleUseCase(perm, permUC, principals),
		usecase.NewRevokeCourseRoleUseCase(perm, permUC, principals),
		usecase.NewListCourseGrantsUseCase(perm, permUC, principals),
		usecase.NewGrantChapterRoleUseCase(perm, permUC, principals),
		usecase.NewRevokeChapterRoleUseCase(perm, permUC, principals),
		usecase.NewListChapterGrantsUseCase(perm, permUC, principals),
		usecase.NewListGrantableMaterialPrincipalsUseCase(perm, permUC, principals),
	)
}
