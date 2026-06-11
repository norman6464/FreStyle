package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// SetCompanyActiveInput は会社アカウントの有効/無効更新の入力。
type SetCompanyActiveInput struct {
	CompanyID uint64
	Active    bool
}

// SetCompanyActiveUseCase は会社アカウントの有効/無効を切り替える（super_admin 用）。
// 無効化すると、その会社の全ユーザーが middleware でログイン/利用を弾かれる。
type SetCompanyActiveUseCase struct {
	companies repository.CompanyRepository
}

// NewSetCompanyActiveUseCase は SetCompanyActiveUseCase を生成する。
func NewSetCompanyActiveUseCase(c repository.CompanyRepository) *SetCompanyActiveUseCase {
	return &SetCompanyActiveUseCase{companies: c}
}

// Execute は会社の有効/無効を更新する。
func (u *SetCompanyActiveUseCase) Execute(ctx context.Context, in SetCompanyActiveInput) error {
	return u.companies.UpdateActive(ctx, in.CompanyID, in.Active)
}
