package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/legacyrepository"
)

type ListCompaniesUseCase struct {
	repo legacyrepository.CompanyRepository
}

func NewListCompaniesUseCase(r legacyrepository.CompanyRepository) *ListCompaniesUseCase {
	return &ListCompaniesUseCase{repo: r}
}

func (u *ListCompaniesUseCase) Execute(ctx context.Context) ([]domain.Company, error) {
	return u.repo.ListAll(ctx)
}
