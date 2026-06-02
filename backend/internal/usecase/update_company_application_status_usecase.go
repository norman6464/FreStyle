package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// UpdateCompanyApplicationStatusUseCase は申請の status を更新する（super_admin 専用）。
type UpdateCompanyApplicationStatusUseCase struct {
	apps repository.CompanyApplicationRepository
}

func NewUpdateCompanyApplicationStatusUseCase(apps repository.CompanyApplicationRepository) *UpdateCompanyApplicationStatusUseCase {
	return &UpdateCompanyApplicationStatusUseCase{apps: apps}
}

// ErrCompanyApplicationBadStatus は許容外の status が渡されたとき返る。
var ErrCompanyApplicationBadStatus = errors.New("company application: invalid status")

func (u *UpdateCompanyApplicationStatusUseCase) Execute(ctx context.Context, id uint64, status string) error {
	if id == 0 {
		return errors.New("id is required")
	}
	// 前後空白・大文字混在を正規化してから検証する。
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case domain.CompanyApplicationStatusPending,
		domain.CompanyApplicationStatusApproved,
		domain.CompanyApplicationStatusRejected:
	default:
		return ErrCompanyApplicationBadStatus
	}
	return u.apps.UpdateStatus(ctx, id, normalized)
}
