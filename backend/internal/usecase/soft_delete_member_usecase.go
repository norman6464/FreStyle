package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// SoftDeleteMemberUseCase は従業員を論理削除する（deleted_at = NOW()）。
// 認可は SetMemberActiveUseCase と同じ（super_admin は全社 / company_admin は自社 / 自分は不可）。
type SoftDeleteMemberUseCase struct {
	users repository.UserRepository
}

func NewSoftDeleteMemberUseCase(u repository.UserRepository) *SoftDeleteMemberUseCase {
	return &SoftDeleteMemberUseCase{users: u}
}

// Execute は actor の権限内で targetUserID を論理削除する。
func (uc *SoftDeleteMemberUseCase) Execute(ctx context.Context, actor *domain.User, targetUserID uint64) error {
	target, err := uc.users.FindByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrMemberNotFound
	}
	if err := authorizeMemberManagement(actor, target); err != nil {
		return err
	}
	return uc.users.SoftDelete(ctx, targetUserID)
}
