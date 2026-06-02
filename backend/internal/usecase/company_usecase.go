package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListCompaniesUseCase は全 company を返す（SuperAdmin 専用）。
type ListCompaniesUseCase struct {
	repo repository.CompanyRepository
}

func NewListCompaniesUseCase(r repository.CompanyRepository) *ListCompaniesUseCase {
	return &ListCompaniesUseCase{repo: r}
}

func (u *ListCompaniesUseCase) Execute(ctx context.Context) ([]domain.Company, error) {
	return u.repo.ListAll(ctx)
}
