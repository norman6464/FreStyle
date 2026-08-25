package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// PromoteCognitoAdminRoleInput は昇格対象を Cognito の subject で指定する。
type PromoteCognitoAdminRoleInput struct {
	CognitoSub string
}

// PromoteCognitoAdminRoleUseCase は Cognito の admin グループに属するユーザーの DB role を
// super_admin へ同期する。昇格だけを行い、剥奪（降格）はしない。
// 呼び元（handler）が repository を直接触らないための境界でもある。
type PromoteCognitoAdminRoleUseCase struct {
	users repository.UserRepository
}

// NewPromoteCognitoAdminRoleUseCase は PromoteCognitoAdminRoleUseCase を生成する。
func NewPromoteCognitoAdminRoleUseCase(
	users repository.UserRepository,
) *PromoteCognitoAdminRoleUseCase {
	return &PromoteCognitoAdminRoleUseCase{users: users}
}

// Execute は subject のユーザーを引き、まだ管理者ロールでなければ super_admin へ昇格する。
// promoted は実際に昇格したか。ユーザーが存在しない・既に管理者ロールなら (false, nil)。
// 失敗は握り潰さずエラーとして返す（呼び元がログに残せるようにする）。
func (u *PromoteCognitoAdminRoleUseCase) Execute(
	ctx context.Context,
	in PromoteCognitoAdminRoleInput,
) (promoted bool, err error) {
	if u.users == nil {
		return false, errors.New("user repository not configured")
	}
	if in.CognitoSub == "" {
		return false, errors.New("cognito sub is empty")
	}

	existing, findErr := u.users.FindByCognitoSub(ctx, in.CognitoSub)
	if findErr != nil {
		// 「DB 障害」と「そのユーザーが居ない」は区別する。呼び元が両者を同じ無反応に
		// 畳むと、恒久的な失敗が誰にも気付かれないまま残る。
		return false, fmt.Errorf("find user by cognito sub: %w", findErr)
	}
	if existing == nil {
		return false, nil
	}
	if existing.Role == domain.RoleSuperAdmin || existing.Role == domain.RoleCompanyAdmin {
		return false, nil
	}

	if err := u.users.UpdateRole(ctx, existing.ID, domain.RoleSuperAdmin); err != nil {
		return false, fmt.Errorf("update role to super admin: %w", err)
	}
	return true, nil
}
