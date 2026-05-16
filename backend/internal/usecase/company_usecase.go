package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ListCompaniesUseCase は 全 company を 返す Use Case。 SuperAdmin 専用 画面 用。
// 依存 port: [repository.CompanyRepository]。
type ListCompaniesUseCase struct {
	repo repository.CompanyRepository
}

func NewListCompaniesUseCase(r repository.CompanyRepository) *ListCompaniesUseCase {
	return &ListCompaniesUseCase{repo: r}
}

func (u *ListCompaniesUseCase) Execute(ctx context.Context) ([]domain.Company, error) {
	return u.repo.ListAll(ctx)
}
