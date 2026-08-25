package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListCompanyMembersUseCase は actor（company_admin）の自社の従業員一覧を返す。
type ListCompanyMembersUseCase struct {
	users repository.UserRepository
}

func NewListCompanyMembersUseCase(u repository.UserRepository) *ListCompanyMembersUseCase {
	return &ListCompanyMembersUseCase{users: u}
}

// Execute は actor の所属会社の従業員一覧を返す。会社未所属なら空。
func (uc *ListCompanyMembersUseCase) Execute(ctx context.Context, actor *domain.User) ([]domain.User, error) {
	if actor == nil {
		return []domain.User{}, nil
	}
	// 未所属（運営管理者など）は「自社」が無いので空を返す。
	companyID, affiliated := actor.CompanyRef().CompanyID()
	if !affiliated {
		return []domain.User{}, nil
	}
	return uc.users.ListByCompanyID(ctx, companyID)
}
