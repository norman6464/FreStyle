package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CourseUseCase はコースの get / create / update / delete を 1 構造体で扱う。
// canManage は teaching_material_usecase と共有。trainee は published のみ閲覧、
// 編集系は同一 company の company_admin または super_admin。Delete は配下教材も cascade 削除。
// 一覧は進捗集計を伴うため ListCoursesWithProgressUseCase が担う(FRESTYLE-98)。
//
//naminglint:allow 複数 CRUD を束ねる集約 usecase のため Execute 単一メソッドではなく Get/Create 等で公開する
type CourseUseCase struct {
	courses   repository.CourseRepository
	materials repository.TeachingMaterialRepository
}

func NewCourseUseCase(courses repository.CourseRepository, materials repository.TeachingMaterialRepository) *CourseUseCase {
	return &CourseUseCase{courses: courses, materials: materials}
}

func (uc *CourseUseCase) Get(ctx context.Context, id uint64, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) (*domain.Course, error) {
	c, err := uc.courses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canReadCourse(c, actorWorkspace, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return c, nil
}

// canReadCourse は対象コースを actorWorkspace が閲覧できるかを判定する。
// FRESTYLE-402（段4横展開）: 対象コースとの比較を company_id 直読みから workspace_id 経由へ
// 切り替え済み。courses.workspace_id は起動時バックフィル + InsertCourse の dual-write
// （FRESTYLE-399）により、リクエストを捌く時点で必ず埋まっているため company_id との
// 併用フォールバックは持たない。
func canReadCourse(c *domain.Course, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) bool {
	if !courseBelongsToWorkspace(c, actorWorkspace, actorRole) {
		return false
	}
	if !c.IsPublished && !canManage(actorRole) {
		return false
	}
	return true
}

// courseBelongsToWorkspace は super_admin か、対象コースが actorWorkspace に属するかを返す。
// Get（canReadCourse 経由）/ Update / Delete で同じ形の所属チェックが個別に書かれていた
// 重複を、この 1 つに集約した（判定結果は変えない）。
func courseBelongsToWorkspace(c *domain.Course, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) bool {
	if actorRole == domain.RoleSuperAdmin {
		return true
	}
	wid, ok := c.WorkspaceRef().WorkspaceID()
	// 未所属の actor・対象コースの workspace_id 未設定はどちらも一致し得ない。
	return ok && actorWorkspace.Matches(wid)
}

type CreateCourseInput struct {
	ActorUserID  uint64
	ActorCompany domain.CompanyRef
	ActorRole    domain.RoleName
	Title        string
	Description  string
	Category     string
	Language     string
	SortOrder    int
	IsPublished  bool
}

func (uc *CourseUseCase) Create(ctx context.Context, in CreateCourseInput) (*domain.Course, error) {
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden: only company_admin or super_admin can create courses")
	}
	// 作成したコースの所属先が決まらないため、未所属の actor は super_admin でも作成できない。
	companyID, affiliated := in.ActorCompany.CompanyID()
	if !affiliated {
		return nil, fmt.Errorf("actor must belong to a company")
	}
	if !domain.IsValidCourseCategory(in.Category) {
		return nil, fmt.Errorf("invalid course category: %s", in.Category)
	}
	c := &domain.Course{
		CompanyID:       companyID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		Description:     in.Description,
		Category:        in.Category,
		Language:        in.Language,
		SortOrder:       in.SortOrder,
		IsPublished:     in.IsPublished,
	}
	if err := uc.courses.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

type UpdateCourseInput struct {
	ID             uint64
	ActorWorkspace domain.WorkspaceRef
	ActorRole      domain.RoleName
	Title          string
	Description    string
	Category       string
	Language       string
	SortOrder      int
	IsPublished    bool
}

func (uc *CourseUseCase) Update(ctx context.Context, in UpdateCourseInput) (*domain.Course, error) {
	existing, err := uc.courses.GetByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	if !courseBelongsToWorkspace(existing, in.ActorWorkspace, in.ActorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	if !domain.IsValidCourseCategory(in.Category) {
		return nil, fmt.Errorf("invalid course category: %s", in.Category)
	}
	existing.Title = in.Title
	existing.Description = in.Description
	existing.Category = in.Category
	existing.Language = in.Language
	existing.SortOrder = in.SortOrder
	existing.IsPublished = in.IsPublished
	if err := uc.courses.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete はコースと配下教材を同時に削除する（cascade 相当）。
func (uc *CourseUseCase) Delete(ctx context.Context, id uint64, actorWorkspace domain.WorkspaceRef, actorRole domain.RoleName) error {
	existing, err := uc.courses.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManage(actorRole) {
		return fmt.Errorf("forbidden")
	}
	if !courseBelongsToWorkspace(existing, actorWorkspace, actorRole) {
		return fmt.Errorf("forbidden")
	}
	if err := uc.materials.DeleteByCourse(ctx, id); err != nil {
		return err
	}
	return uc.courses.Delete(ctx, id)
}
