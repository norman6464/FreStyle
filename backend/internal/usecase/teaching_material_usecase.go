package usecase

import (
	"context"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// TeachingMaterialUseCase は教材機能の 5 操作（list / get / create / update / delete）を
// まとめて 1 構造体で扱う。 各操作は独立した Execute メソッドではなく
// 名前付きメソッドにすることで、 handler 側のディスパッチを単純化する。
//
// アクセス制御:
//   - List: 同一 company の company_admin / trainee / super_admin
//   - trainee は published のみ、 company_admin / super_admin は draft 含む全件
//   - Get: 同一 company か super_admin
//   - trainee は published のみ閲覧可
//   - Create / Update / Delete: 同一 company の company_admin、 または super_admin
type TeachingMaterialUseCase struct {
	repo repository.TeachingMaterialRepository
}

func NewTeachingMaterialUseCase(repo repository.TeachingMaterialRepository) *TeachingMaterialUseCase {
	return &TeachingMaterialUseCase{repo: repo}
}

// canManage は教材を作成 / 編集 / 削除できる role 判定。
func canManage(role string) bool {
	return role == domain.RoleCompanyAdmin || role == domain.RoleSuperAdmin
}

// List は actor の role / company に応じて閲覧可能な教材一覧を返す。
//   - companyID=0 のユーザ（super_admin で会社未紐付け等）は空配列を返す
//   - super_admin かつ companyID=0 でも、 トップで会社全体管理する画面は別途用意する想定で、
//     ここでは「自分の company に紐付く教材だけ」に絞る
func (uc *TeachingMaterialUseCase) List(ctx context.Context, actorCompanyID uint64, actorRole string) ([]domain.TeachingMaterial, error) {
	if actorCompanyID == 0 {
		return []domain.TeachingMaterial{}, nil
	}
	includeUnpublished := canManage(actorRole)
	return uc.repo.ListByCompany(ctx, actorCompanyID, includeUnpublished)
}

// Get は ID 指定で 1 件取得する。 actor の company と一致しないと 403 相当（nil）。
func (uc *TeachingMaterialUseCase) Get(ctx context.Context, id, actorCompanyID uint64, actorRole string) (*domain.TeachingMaterial, error) {
	m, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !canRead(m, actorCompanyID, actorRole) {
		return nil, fmt.Errorf("forbidden")
	}
	return m, nil
}

func canRead(m *domain.TeachingMaterial, actorCompanyID uint64, actorRole string) bool {
	if actorRole == domain.RoleSuperAdmin {
		return true
	}
	if m.CompanyID != actorCompanyID {
		return false
	}
	if !m.IsPublished && !canManage(actorRole) {
		return false
	}
	return true
}

// CreateInput は POST 入力。
type CreateTeachingMaterialInput struct {
	ActorUserID    uint64
	ActorCompanyID uint64
	ActorRole      string
	Title          string
	Content        string
	IsPublished    bool
}

func (uc *TeachingMaterialUseCase) Create(ctx context.Context, in CreateTeachingMaterialInput) (*domain.TeachingMaterial, error) {
	if !canManage(in.ActorRole) {
		return nil, fmt.Errorf("forbidden: only company_admin or super_admin can create materials")
	}
	if in.ActorCompanyID == 0 {
		return nil, fmt.Errorf("actor must belong to a company")
	}
	m := &domain.TeachingMaterial{
		CompanyID:       in.ActorCompanyID,
		CreatedByUserID: in.ActorUserID,
		Title:           in.Title,
		Content:         in.Content,
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
