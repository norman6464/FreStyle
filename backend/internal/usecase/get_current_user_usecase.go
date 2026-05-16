package usecase

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetCurrentUserUseCase は Cognito sub から現在のユーザー情報を引く。
// 依存 port: [repository.UserRepository] (FindByCognitoSub を 使用)。
type GetCurrentUserUseCase struct {
	users repository.UserRepository
}

func NewGetCurrentUserUseCase(users repository.UserRepository) *GetCurrentUserUseCase {
	return &GetCurrentUserUseCase{users: users}
}

func (u *GetCurrentUserUseCase) Execute(ctx context.Context, cognitoSub string) (*domain.User, error) {
	return u.users.FindByCognitoSub(ctx, cognitoSub)
}
