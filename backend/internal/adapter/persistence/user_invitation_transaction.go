package persistence

import (
	"context"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

type userInvitationTransactionRunner struct {
	db *gorm.DB
}

// NewUserInvitationTransactionRunner はユーザーと招待の
// トランザクション実装を生成する。
func NewUserInvitationTransactionRunner(
	db *gorm.DB,
) repository.UserInvitationTransactionRunner {
	return &userInvitationTransactionRunner{db: db}
}

func (r *userInvitationTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(
		users repository.UserRepository,
		invitations repository.AdminInvitationRepository,
	) error,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(
			NewUserRepository(tx),
			NewAdminInvitationRepository(tx),
		)
	})
}
