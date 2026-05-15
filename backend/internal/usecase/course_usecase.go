package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CourseUseCase はコース機能の操作（list / get / create / update / delete）を
// 1 構造体で扱う。 内部ヘルパは teaching_material_usecase の canManage を共有する。
//
// アクセス制御:
//   - List: 同一 company の company_admin / trainee / super_admin
//     trainee は published のみ、 company_admin / super_admin は draft 含む全件
//   - Get: 同一 company か super_admin（trainee は published のみ）
//   - Create / Update / Delete: 同一 company の company_admin、 または super_admin
//   - Delete: コース内の教材も同時に削除する（cascade）
type CourseUseCase struct {
	courses   repository.CourseRepository
	materials repository.TeachingMaterialRepository
}

func NewCourseUseCase(courses repository.CourseRepository, materials repository.TeachingMaterialRepository) *CourseUseCase {
	return &CourseUseCase{courses: courses, materials: materials}
}

func (uc *CourseUseCase) List(ctx context.Context, actorCompanyID uint64, actorRole string) ([]domain.Course, error) {
	if actorCompanyID == 0 {
		return []domain.Course{}, nil
	}
	includeUnpublished := canManage(actorRole)
	return uc.courses.ListByCompany(ctx, actorCompanyID, includeUnpublished)
}

func (uc *CourseUseCase) Get(ctx context.Context, id, actorCompanyID uint64, actorRole string) (*domain.Course, error) {
	c, err := uc.courses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canReadCourse(c, actorCompanyID, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return c, nil
}

func canReadCourse(c *domain.Course, actorCompanyID uint64, actorRole string) bool {
	if actorRole == domain.RoleSuperAdmin {
		return true
	}
	if c.CompanyID != actorCompanyID {
		return false
	}
	if !c.IsPublished && !canManage(actorRole) {
		return false
	}
	return true
}

type CreateCourseInput struct {
	ActorUserID    uint64
	ActorCompanyID uint64
	ActorRole      string
	Title          string
	Description    string
	SortOrder      int
	IsPublished    bool
}

func (uc *CourseUseCase) Create(ctx context.Context, in CreateCourseInput) (*domain.Course, error) {
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden: only company_admin or super_admin can create courses")
	}
	if in.ActorCompanyID == 0 {
		return nil, fmt.Errorf("actor must belong to a company")
	}
	c := &domain.Course{
		CompanyID:       in.ActorCompanyID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		Description:     in.Description,
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
	ActorCompanyID uint64
	ActorRole      string
	Title          string
	Description    string
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
	if in.ActorRole != domain.RoleSuperAdmin && existing.CompanyID != in.ActorCompanyID {
		return nil, fmt.Errorf("forbidden")
	}
	existing.Title = in.Title
	existing.Description = in.Description
	existing.SortOrder = in.SortOrder
	existing.IsPublished = in.IsPublished
	if err := uc.courses.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete はコースとその配下の教材を同時に削除する。 教材は repository の
// DeleteByCourse で一括削除（cascade 相当）。
func (uc *CourseUseCase) Delete(ctx context.Context, id, actorCompanyID uint64, actorRole string) error {
	existing, err := uc.courses.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManage(actorRole) {
		return fmt.Errorf("forbidden")
	}
	if actorRole != domain.RoleSuperAdmin && existing.CompanyID != actorCompanyID {
		return fmt.Errorf("forbidden")
	}
	if err := uc.materials.DeleteByCourse(ctx, id); err != nil {
		return err
	}
	return uc.courses.Delete(ctx, id)
}
