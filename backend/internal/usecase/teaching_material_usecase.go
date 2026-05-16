package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// TeachingMaterialUseCase は教材機能の 5 操作（list / get / create / update / delete）を
// まとめて 1 構造体で扱う。 各操作は独立した Execute メソッドではなく
// 名前付きメソッドにすることで、 handler 側のディスパッチを単純化する。
//
// 教材は必ず Course に所属する前提のため、 list は Course 単位、 create は course_id 必須。
//
// アクセス制御:
//   - List: 同一 company の company_admin / trainee / super_admin
//     trainee は published のみ、 company_admin / super_admin は draft 含む全件
//     trainee は所属コースが is_published=false の場合は閲覧不可
//   - Get: 同一 company か super_admin、 trainee は published のみ閲覧可
//   - Create / Update / Delete: 同一 company の company_admin、 または super_admin
//
// 依存 port: [repository.TeachingMaterialRepository] + [repository.CourseRepository]
// (course との 整合性 検証 用)。 章 003 で 全 体 walk-through、 章 022 で 詳細 解説。
type TeachingMaterialUseCase struct {
	repo    repository.TeachingMaterialRepository
	courses repository.CourseRepository
}

func NewTeachingMaterialUseCase(repo repository.TeachingMaterialRepository, courses repository.CourseRepository) *TeachingMaterialUseCase {
	return &TeachingMaterialUseCase{repo: repo, courses: courses}
}

// canManage は教材を作成 / 編集 / 削除できる role 判定。
func canManage(role string) bool {
	return role == domain.RoleCompanyAdmin || role == domain.RoleSuperAdmin
}

// List は backward-compat 用。 既存 frontend が GET /teaching-materials を直接叩くため、
// company 内の全教材を返す。 frontend がコース対応に切り替わったら削除予定。
func (uc *TeachingMaterialUseCase) List(ctx context.Context, actorCompanyID uint64, actorRole string) ([]domain.TeachingMaterial, error) {
	if actorCompanyID == 0 {
		return []domain.TeachingMaterial{}, nil
	}
	includeUnpublished := canManage(actorRole)
	return uc.repo.ListByCompany(ctx, actorCompanyID, includeUnpublished)
}

// ListByCourse は指定コース配下の教材を返す。 actor の role / company を検証してから
// repository を呼ぶ。 trainee はコース自体が is_published=true の場合のみ閲覧可。
func (uc *TeachingMaterialUseCase) ListByCourse(ctx context.Context, courseID, actorCompanyID uint64, actorRole string) ([]domain.TeachingMaterial, error) {
	course, err := uc.courses.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if !canReadCourse(course, actorCompanyID, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	includeUnpublished := canManage(actorRole)
	return uc.repo.ListByCourse(ctx, courseID, includeUnpublished)
}

// Get は ID 指定で 1 件取得する。 actor の company と一致しないと 403 相当（nil）。
// 所属コース自体が trainee に対して非公開なら閲覧不可。
func (uc *TeachingMaterialUseCase) Get(ctx context.Context, id, actorCompanyID uint64, actorRole string) (*domain.TeachingMaterial, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	course, err := uc.courses.GetByID(ctx, m.CourseID)
	if err != nil {
		return nil, err
	}
	if !canRead(m, course, actorCompanyID, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return m, nil
}

func canRead(m *domain.TeachingMaterial, course *domain.Course, actorCompanyID uint64, actorRole string) bool {
	if actorRole == domain.RoleSuperAdmin {
		return true
	}
	if m.CompanyID != actorCompanyID {
		return false
	}
	// 所属コースも閲覧可能でないと教材も見せない（trainee は published コース内の published 教材のみ）。
	if !canReadCourse(course, actorCompanyID, actorRole) {
		return false
	}
	if !m.IsPublished && !canManage(actorRole) {
		return false
	}
	return true
}

// CreateInput は POST 入力。 CourseID は必須（所属コース）。
type CreateTeachingMaterialInput struct {
	ActorUserID    uint64
	ActorCompanyID uint64
	ActorRole      string
	CourseID       uint64
	Title          string
	Content        string
	OrderInCourse  int
	IsPublished    bool
}

func (uc *TeachingMaterialUseCase) Create(ctx context.Context, in CreateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden: only company_admin or super_admin can create materials")
	}
	if in.ActorCompanyID == 0 {
		return nil, fmt.Errorf("actor must belong to a company")
	}
	if in.CourseID == 0 {
		return nil, fmt.Errorf("course_id is required")
	}
	// コースの存在 / company 一致を確認する。
	course, err := uc.courses.GetByID(ctx, in.CourseID)
	if err != nil {
		return nil, err
	}
	if in.ActorRole != domain.RoleSuperAdmin && course.CompanyID != in.ActorCompanyID {
		return nil, fmt.Errorf("forbidden")
	}
	m := &domain.TeachingMaterial{
		CompanyID:       course.CompanyID,
		CourseID:        in.CourseID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		Content:         in.Content,
		OrderInCourse:   in.OrderInCourse,
		IsPublished:     in.IsPublished,
	}
	if err := uc.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

type UpdateTeachingMaterialInput struct {
	ID             uint64
	ActorCompanyID uint64
	ActorRole      string
	Title          string
	Content        string
	OrderInCourse  int
	IsPublished    bool
}

func (uc *TeachingMaterialUseCase) Update(ctx context.Context, in UpdateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	existing, err := uc.repo.GetByID(ctx, in.ID)
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
	existing.Content = in.Content
	existing.OrderInCourse = in.OrderInCourse
	existing.IsPublished = in.IsPublished
	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *TeachingMaterialUseCase) Delete(ctx context.Context, id, actorCompanyID uint64, actorRole string) error {
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if !canManage(actorRole) {
		return fmt.Errorf("forbidden")
	}
	if actorRole != domain.RoleSuperAdmin && existing.CompanyID != actorCompanyID {
		return fmt.Errorf("forbidden")
	}
	return uc.repo.Delete(ctx, id)
}
