package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListCompanyApplicationsUseCase は全企業申請を新しい順で返す（super_admin 専用）。
type ListCompanyApplicationsUseCase struct {
	apps repository.CompanyApplicationRepository
}

func NewListCompanyApplicationsUseCase(apps repository.CompanyApplicationRepository) *ListCompanyApplicationsUseCase {
	return &ListCompanyApplicationsUseCase{apps: apps}
}

func (u *ListCompanyApplicationsUseCase) Execute(ctx context.Context) ([]domain.CompanyApplication, error) {
	return u.apps.ListAll(ctx)
}
