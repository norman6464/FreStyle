package usecase

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

var (
	// ErrCannotManageSelf は自分自身を無効化/削除しようとしたときのエラー（自己ロックアウト防止）。
	ErrCannotManageSelf = errors.New("cannot disable or delete yourself")
	// ErrMemberNotFound は対象 user が存在しない（404 相当）。
	ErrMemberNotFound = errors.New("member not found")
)

// authorizeMemberManagement は actor が target を管理（停止/削除）できるかを判定する。
// super_admin は全社の user を、company_admin は自社の user のみ操作できる。自分自身は不可。
func authorizeMemberManagement(actor, target *domain.User) error {
	if actor == nil || target == nil {
		return ErrMemberNotInActorCompany
	}
	if actor.ID == target.ID {
		return ErrCannotManageSelf
	}
	if actor.Role == domain.RoleSuperAdmin {
		return nil
	}
	// company_admin は自社（同一 company_id）の user のみ。
	if actor.Role != domain.RoleCompanyAdmin || actor.CompanyID == nil ||
		target.CompanyID == nil || *actor.CompanyID != *target.CompanyID {
		return ErrMemberNotInActorCompany
	}
	return nil
}

// SetMemberActiveUseCase は従業員アカウントの有効/無効を切り替える（停止/再開）。
type SetMemberActiveUseCase struct {
	users repository.UserRepository
}

func NewSetMemberActiveUseCase(u repository.UserRepository) *SetMemberActiveUseCase {
	return &SetMemberActiveUseCase{users: u}
}

// Execute は actor の権限内で targetUserID の有効/無効を更新する。
func (uc *SetMemberActiveUseCase) Execute(ctx context.Context, actor *domain.User, targetUserID uint64, active bool) error {
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
	return uc.users.UpdateActive(ctx, targetUserID, active)
}
